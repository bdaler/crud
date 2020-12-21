package managers

import (
	"context"
	"github.com/bdaler/crud/pkg/types"
	"github.com/bdaler/crud/pkg/utils"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{pool: db}
}

func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	sqlStatement := `SELECT manager_id FROM managers_tokens WHERE token = $1`
	err := s.pool.QueryRow(ctx, sqlStatement, token).Scan(&id)
	if err != nil {
		log.Print(err)
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, nil
	}

	return id, nil
}

func (s *Service) IsAdmin(ctx context.Context, id int64) (isAdmin bool) {
	sqlStmt := `SELECT is_admin FROM managers WHERE id = $1`
	err := s.pool.QueryRow(ctx, sqlStmt, id).Scan(&isAdmin)
	if err != nil {
		return false
	}

	return
}

func (s *Service) Create(ctx context.Context, item *types.Manager) (string, error) {
	var token string
	var id int64
	sqlStmt := `INSERT INTO managers(name,phone,is_admin) VALUES ($1,$2,$3) ON CONFLICT (phone) DO NOTHING RETURNING id;`
	err := s.pool.QueryRow(ctx, sqlStmt, item.Name, item.Phone, item.IsAdmin).Scan(&id)
	if err != nil {
		log.Print(err)
		return "", types.ErrInternal
	}

	token, err = utils.GenerateTokenStr()
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `INSERT INTO managers_tokens(token,manager_id) VALUES($1,$2)`, token, id)
	if err != nil {
		return "", types.ErrInternal
	}

	return token, nil
}

func (s *Service) Token(ctx context.Context, phone, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(
		ctx,
		`SELECT id, password FROM managers WHERE phone = $1`,
		phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", types.ErrInvalidPassword
	}
	if err != nil {
		return "", types.ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", types.ErrInvalidPassword
	}

	token, err = utils.GenerateTokenStr()
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(
		ctx, `INSERT INTO managers_tokens(token,manager_id) VALUES ($1,$2)`,
		token,
		id)
	if err != nil {
		return "", types.ErrInternal
	}

	return token, nil
}

func (s *Service) SaveProduct(ctx context.Context, product *types.Product) (*types.Product, error) {
	var err error

	if product.ID == 0 {
		sql := `INSERT INTO products(name,qty,price) VALUES ($1,$2,$3) RETURNING id, name, qty, price, active, created;`
		err = s.pool.QueryRow(
			ctx,
			sql,
			product.Name,
			product.Qty,
			product.Price).
			Scan(&product.ID,
				&product.Name,
				&product.Qty,
				&product.Price,
				&product.Active,
				&product.Created)
	} else {
		sql := `UPDATE products SET name=$1, qty=$2,price=$3 WHERE id = $4 RETURNING id, name, qty, price, active, created;`
		err = s.pool.QueryRow(
			ctx,
			sql,
			product.Name,
			product.Qty,
			product.Price,
			product.ID).
			Scan(&product.ID,
				&product.Name,
				&product.Qty,
				&product.Price,
				&product.Active,
				&product.Created)
	}

	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}
	return product, nil
}

func (s *Service) MakeSalePosition(ctx context.Context, position *types.SalePosition) bool {
	active := false
	qty := 0
	if err := s.pool.QueryRow(
		ctx, `SELECT qty,active FROM products WHERE id = $1`, position.ProductID).
		Scan(&qty, &active); err != nil {
		return false
	}
	if qty < position.Qty || !active {
		return false
	}
	if _, err := s.pool.Exec(
		ctx, `UPDATE products SET qty = $1 WHERE id = $2`,
		qty-position.Qty,
		position.ProductID); err != nil {
		log.Print(err)
		return false
	}

	return true
}

func (s *Service) MakeSale(ctx context.Context, sale *types.Sale) (*types.Sale, error) {
	positionsSql := "INSERT INTO sales_positions (sale_id,product_id,qty,price) VALUES "
	sql := `INSERT INTO sales(manager_id,customer_id) VALUES ($1,$2) RETURNING id, created;`
	err := s.pool.QueryRow(
		ctx,
		sql,
		sale.ManagerID,
		sale.CustomerID).Scan(&sale.ID, &sale.Created)
	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}
	for _, position := range sale.Positions {
		if !s.MakeSalePosition(ctx, position) {
			log.Print("Invalid position")
			return nil, types.ErrInternal
		}
		positionsSql += "(" + strconv.FormatInt(sale.ID, 10) + "," + strconv.FormatInt(position.ProductID, 10) + "," + strconv.Itoa(position.Price) + "," + strconv.Itoa(position.Qty) + "),"
	}

	positionsSql = positionsSql[0 : len(positionsSql)-1]

	log.Print(positionsSql)
	_, err = s.pool.Exec(ctx, positionsSql)
	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}

	return sale, nil
}

func (s *Service) GetSales(ctx context.Context, id int64) (sum int, err error) {

	sqlstmt := `
	SELECT COALESCE(SUM(sp.qty * sp.price),0) total
	FROM managers m
	LEFT JOIN sales s ON s.manager_id= $1
	LEFT JOIN sales_positions sp ON sp.sale_id = s.id
	GROUP BY m.id
	LIMIT 1`

	err = s.pool.QueryRow(ctx, sqlstmt, id).Scan(&sum)
	if err != nil {
		log.Print(err)
		return 0, types.ErrInternal
	}
	return sum, nil
}

func (s *Service) Products(ctx context.Context) ([]*types.Product, error) {
	items := make([]*types.Product, 0)
	sql := `SELECT id, name, price, qty FROM products WHERE active = TRUE ORDER BY ID LIMIT 500`
	rows, err := s.pool.Query(ctx, sql)

	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, types.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &types.Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Service) RemoveProductByID(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		log.Print(err)
		return types.ErrInternal
	}

	return nil
}

func (s *Service) RemoveCustomerByID(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		log.Print(err)
		return types.ErrInternal
	}

	return nil
}

func (s *Service) Customers(ctx context.Context) ([]*types.Customer, error) {
	items := make([]*types.Customer, 0)
	sql := `SELECT id, name, phone, active, created FROM customers WHERE active = TRUE ORDER BY ID LIMIT 500`
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, types.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &types.Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Service) ChangeCustomer(ctx context.Context, customer *types.Customer) (*types.Customer, error) {
	sql := `UPDATE customers SET name = $2, phone = $3, active = $4  WHERE id = $1 RETURNING name, phone, active`
	if err := s.pool.QueryRow(
		ctx,
		sql,
		customer.ID,
		customer.Name,
		customer.Phone,
		customer.Active).
		Scan(&customer.Name, &customer.Phone, &customer.Active); err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}

	return customer, nil
}
