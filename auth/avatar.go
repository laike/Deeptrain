package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const avatarSize = 128

func getBackgroundColor() color.RGBA {
	return color.RGBA{R: uint8(rand.Intn(200)), G: uint8(rand.Intn(200)), B: uint8(rand.Intn(200)), A: 255}
}

func SaveAvatar(username string) {
	first, _ := utf8.DecodeRuneInString(username) // Extract the first rune from the username (UTF-8 safe)

	background := getBackgroundColor()

	var canvas strings.Builder
	canvas.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, avatarSize, avatarSize))

	canvas.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" style="fill:rgb(%d,%d,%d);" />`, avatarSize, avatarSize, background.R, background.G, background.B))
	canvas.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" dominant-baseline="middle" font-family="Arial" font-size="50" fill="white">%s</text>`, avatarSize/2, avatarSize/2, string(first)))
	canvas.WriteString("</svg>")

	data := []byte(canvas.String())
	if err := os.WriteFile(fmt.Sprintf("storage/avatar/%s.svg", username), data, 0644); err != nil {
		log.Println(err)
	} else {
		SaveAvatarConfig(username, fmt.Sprintf("%s.svg", username))
	}
}

func SaveAvatarConfig(username, path string) bool {
	if _, err := os.Stat("storage/avatar/config.json"); os.IsNotExist(err) {
		data := []byte("{}")
		if err := os.WriteFile("storage/avatar/config.json", data, 0644); err != nil {
			log.Println(err)
			return false
		}
	}

	var config map[string]string

	data, err := os.ReadFile("storage/avatar/config.json")
	if err != nil {
		log.Println(err)
		return false
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Println(err)
		return false
	}

	config[username] = path

	data, err = json.Marshal(config)
	if err != nil {
		log.Println(err)
		return false
	}

	if err := os.WriteFile("storage/avatar/config.json", data, 0644); err != nil {
		log.Println(err)
		return false
	}

	return true
}

func GetAvatarConfigWithCache(c *gin.Context, username string) string {
	cache := c.MustGet("cache").(*redis.Client)
	if value, err := cache.Get(c, fmt.Sprintf("avatar:%s", username)).Result(); err == nil {
		return value
	}

	var config map[string]string
	data, err := os.ReadFile("storage/avatar/config.json")
	if err != nil {
		log.Println(err)
		return ""
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Println(err)
		return ""
	}

	source := config[username]
	if len(source) != 0 {
		cache.Set(c, fmt.Sprintf("avatar:%s", username), source, time.Minute*30)
	}
	return source
}

func GetAvatarView(c *gin.Context) {
	username := c.Param("username")
	if len(username) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	path := GetAvatarConfigWithCache(c, username)
	if len(path) == 0 {
		db := c.MustGet("db").(*sql.DB)
		if IsUserExists(db, username) {
			SaveAvatar(username)
			path = GetAvatarConfigWithCache(c, username)
			c.File(path)
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "avatar not found"})
			return
		}
	}

	c.File(fmt.Sprintf("storage/avatar/%s", path))
}

func PostAvatarView(c *gin.Context) {
	username := c.MustGet("user").(string)
	if len(username) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "username is required"})
		return
	}

	user := User{Username: username}
	db := c.MustGet("db").(*sql.DB)
	if !user.IsActive(db) {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "user is not active"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
		return
	}

	// Validate file format
	valid, ext := isValidImageFormat(file.Filename)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "invalid image format."})
		return
	}

	// Validate file size (2 MiB max)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": "image size should not exceed 2MB"})
		return
	}

	name := fmt.Sprintf("%s.%s", username, ext)
	if err := c.SaveUploadedFile(file, fmt.Sprintf("storage/avatar/%s", name)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "error": err.Error()})
		return
	}

	SaveAvatarConfig(username, name)

	cache := c.MustGet("cache").(*redis.Client)
	cache.Set(c, fmt.Sprintf("avatar:%s", username), name, time.Minute*30)

	c.JSON(http.StatusOK, gin.H{"status": true})
}

func isValidImageFormat(filename string) (bool, string) {
	allowedFormats := map[string]bool{
		"webp": true,
		"png":  true,
		"jpg":  true,
		"jpeg": true,
		"gif":  true,
		"bmp":  true,
		"svg":  true,
	}

	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	return allowedFormats[ext], ext
}
