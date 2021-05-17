package db

import (
	apiModel "github.com/anchamber/genetics-api/model"
	"github.com/anchamber/genetics-tank/db/model"
)

type Options struct {
	Pageination *apiModel.Pageination
	Filters     []*apiModel.Filter
}

type TankDB interface {
	Select(Options) ([]*model.Tank, error)
	SelectByNumber(number uint32) (*model.Tank, error)
	Insert(tank *model.Tank) error
	Update(tank *model.Tank) error
	Delete(number uint32) error
}

type ErrorCode string

const (
	TankAlreadyExists ErrorCode = "tank already exists"
	Unknown             ErrorCode = "unknown error with db occurred"
)
