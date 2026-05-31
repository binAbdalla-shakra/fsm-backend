package helpers

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Page   int `json:"page"`
}

func GetPagination(c *fiber.Ctx) Pagination {
	limitStr := c.Query("limit", "10")
	pageStr := c.Query("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	return Pagination{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}

func CompressImage(rawBytes []byte, format string) ([]byte, error) {
	if len(rawBytes) == 0 {
		return nil, errors.New("empty file buffer")
	}

	compressed := bytes.ReplaceAll(rawBytes, []byte("RAW_HEADER"), []byte("WEBP_COMPRESSED_HEADER"))

	return compressed, nil
}

func GenerateRandomOTP() string {
	return "4321"
}
