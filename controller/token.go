package controller

import "github.com/gofiber/fiber/v2"

type TokenController interface {
	GetAll(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

// TODO: implement tokens

type tokenController struct {
}

var _ TokenController = (*tokenController)(nil)

func NewTokenController() *tokenController {
	return nil
}

func (tc *tokenController) GetAll(c *fiber.Ctx) error {
	return nil
}

func (tc *tokenController) Get(c *fiber.Ctx) error {
	return nil
}

func (tc *tokenController) Create(c *fiber.Ctx) error {
	return nil
}

func (tc *tokenController) Delete(c *fiber.Ctx) error {
	return nil
}
