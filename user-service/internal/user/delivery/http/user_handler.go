package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/virhanali/filmnesia/user-service/internal/user/domain"
	"github.com/virhanali/filmnesia/user-service/internal/user/usecase"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

func NewUserHandler(userUC usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUC,
	}
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("/register", h.Register)
		userGroup.GET("/:id", h.GetUserByID)
		userGroup.GET("/email/:email", h.GetUserByEmail)
		userGroup.GET("/username/:username", h.GetUserByUsername)
		userGroup.PUT("/:id", h.UpdateUser)
		userGroup.DELETE("/:id", h.DeleteUser)
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req domain.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResponse, err := h.userUsecase.Register(c.Request.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrEmailExists:
			c.JSON(http.StatusConflict, gin.H{"error": usecase.ErrEmailExists.Error()})
		case usecase.ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{"error": usecase.ErrUsernameExists.Error()})
		case usecase.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrInvalidInput.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	userResponse, err := h.userUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengguna berdasarkan ID"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	emailParam := c.Param("email")
	if emailParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter cannot be empty"})
		return
	}

	userResponse, err := h.userUsecase.GetByEmail(c.Request.Context(), emailParam)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		case usecase.ErrInvalidInput: // Jika usecase melakukan validasi format email
			c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrInvalidInput.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data by email"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	usernameParam := c.Param("username")
	if usernameParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username parameter cannot be empty"})
		return
	}

	userResponse, err := h.userUsecase.GetByUsername(c.Request.Context(), usernameParam)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		case usecase.ErrInvalidInput: // Jika usecase melakukan validasi format username
			c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrInvalidInput.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data by username"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Tambahkan logika otorisasi di middleware atau di awal usecase
	// Misalnya: Memastikan pengguna yang login adalah userID atau admin.

	userResponse, err := h.userUsecase.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		case usecase.ErrEmailExists:
			c.JSON(http.StatusConflict, gin.H{"error": usecase.ErrEmailExists.Error()})
		case usecase.ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{"error": usecase.ErrUsernameExists.Error()})
		case usecase.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrInvalidInput.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// TODO: Tambahkan logika otorisasi di middleware atau di awal usecase
	// Misalnya: Memastikan pengguna yang login adalah userID atau admin.

	err = h.userUsecase.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User successfully deleted"})
}
