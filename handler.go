package main

import (
	"context"
	product "github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/product"
	user "github.com/daheishandemao/Tiktok-E-commerce/kitex_gen/user"
)

// OrderServiceImpl implements the last service interface defined in the IDL.
type OrderServiceImpl struct{}

// UpdateStatus implements the OrderServiceImpl interface.
func (s *OrderServiceImpl) UpdateStatus(ctx context.Context, orderID string, status string) (resp string, err error) {
	// TODO: Your code here...
	return
}

// GetUserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserInfo(ctx context.Context, userId int64) (resp *user.UserInfo, err error) {
	// TODO: Your code here...
	return
}

// RegisterUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) RegisterUser(ctx context.Context, req *user.RegisterRequest) (resp int64, err error) {
	// TODO: Your code here...
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginRequest) (resp string, err error) {
	// TODO: Your code here...
	return
}

// CheckToken implements the UserServiceImpl interface.
func (s *UserServiceImpl) CheckToken(ctx context.Context, token string) (resp bool, err error) {
	// TODO: Your code here...
	return
}

// GetProduct implements the ProductServiceImpl interface.
func (s *ProductServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (resp *product.ProductInfo, err error) {
	// TODO: Your code here...
	return
}

// DecreaseStock implements the ProductServiceImpl interface.
func (s *ProductServiceImpl) DecreaseStock(ctx context.Context, req *product.DecreaseStockReq) (resp bool, err error) {
	// TODO: Your code here...
	return
}
