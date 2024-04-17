package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", nil)
	})
	app.Get("/stream", streamVideo)
	if err := app.Listen(":9999"); err != nil {
		log.Fatalf("unable to connect with server : %v", err)
	}
}
func streamVideo(c *fiber.Ctx) error {
	filePath := "cut.mp4"

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening video file: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server error")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file information: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server error")
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	filesize := fileInfo.Size()

	// Set Content-Type header
	c.Set(fiber.HeaderContentType, mimeType)
	

	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		// Handle range request
		parts := strings.Split(rangeHeader, "=")
		rangeParts := strings.Split(parts[1], "-")

		start, err := strconv.ParseInt(rangeParts[0], 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid range")
		}

		var end int64
		if len(rangeParts) > 1 && rangeParts[1] != "" {
			end, err = strconv.ParseInt(rangeParts[1], 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid range")
			}
		} else {
			end = filesize - 1
		}

		if start > end || end >= filesize {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid range")
		}

		c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, filesize))
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(end-start+1, 10))
		c.Status(fiber.StatusPartialContent)

		_, copyErr := io.CopyN(c.Response().BodyWriter(), file, end-start+1)
		if copyErr != nil {
			log.Println("Error copying range to response:", copyErr)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
	} else {
		// If no Range header is present, serve the entire video
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(filesize, 10))
		c.Status(fiber.StatusOK)
		_, copyErr := io.Copy(c.Response().BodyWriter(), file)
		if copyErr != nil {
			log.Println("Error copying entire file to response:", copyErr)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
	}

	return nil
}
