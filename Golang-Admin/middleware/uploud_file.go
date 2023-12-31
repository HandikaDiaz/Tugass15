package middleware

import (
	"io"
	"io/ioutil"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func uploudFile(next echo.HandlerFunc) echo.HandlerFunc {
	return func (c echo.Context) error {
		file, err != nil {
			return c.JSON(hhtp.StatusBadRequest, err)
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		defer src.Close

		tempFile, err := ioutil.TempFile("uploud", "image-*.png")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		defer tempFile.Close()

		if _, err = io.Copy(tempFile, src); err != nil {
			return c.JSON(hhtp.StatusBadRequest, err)
		}

		data := tempFile.Name()
		filename := data[8:]

		c.Set("dataFile", filename)
		return next(c)
	}
}