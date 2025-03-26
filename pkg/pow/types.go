package pow

import (
	"fmt"
	"strconv"
	"strings"
)

type Challenge struct {
	Difficulty uint8
	Timestamp  int64
	Resource   string
	Salt       string
}

type Solution struct {
	Ch    Challenge
	Nonce uint64
}

type POWService interface {
	GenerateChallenge(resource string) (*Challenge, error)

	VerifySolution(solution Solution) (bool, error)
}

func (c *Challenge) Stamp(nonce uint64) string {
	return fmt.Sprintf(
		"1:%d:%d:%s:%s:%d",
		c.Timestamp,
		c.Difficulty,
		c.Resource,
		c.Salt,
		nonce,
	)
}

func Parse(stamp string) (*Challenge, uint64, error) {
	c := &Challenge{}
	parts := strings.Split(stamp, ":")
	if len(parts) != 6 {
		return nil, 0, fmt.Errorf("invalid stamp format")
	}

	var err error
	c.Timestamp, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid stamp format")
	}

	difficulty, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid stamp format")
	}

	c.Difficulty = uint8(difficulty)

	c.Resource = parts[3]
	c.Salt = parts[4]

	nonce, err := strconv.ParseUint(parts[5], 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid stamp format")
	}

	return c, nonce, nil
}
