package util

import (
	"fmt"
	"time"

	"github.com/sony/sonyflake"
)

type OrderNoGenerator interface {
	Generate() string
}

type sonyflakeGenerator struct {
	sf *sonyflake.Sonyflake
}

func NewSonyflakeGenerator() OrderNoGenerator {
	return &sonyflakeGenerator{
		sf: sonyflake.NewSonyflake(sonyflake.Settings{
			StartTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}),
	}
}

func (g *sonyflakeGenerator) Generate() string {
	id, _ := g.sf.NextID()
	return fmt.Sprintf("ORD%013d%04d", 
		time.Now().UnixMilli()%1000000000000, 
		id%10000)
}