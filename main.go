package main

import (
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./templates", ".tpl")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", nil)
	})
	app.Get("/stream", streamVideo)
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("unable to connect with server : %v", err)
	}
}
func streamVideo(c *fiber.Ctx) error {
	filePath := "video.mp4"

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
		// ...
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
