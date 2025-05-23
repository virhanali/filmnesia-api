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
		userGroup.POST("/login", h.Login)
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
		case usecase.ErrInvalidInput:
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
	targetUserID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format in URL"})
		return
	}

	authUserIDValue, exists := c.Get(AuthUserIDKey)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get authenticated user ID from context"})
		return
	}
	authUserID, ok := authUserIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authenticated user ID in context is of invalid type"})
		return
	}

	authUserRoleValue, roleExists := c.Get(AuthUserRoleKey)
	if !roleExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User role not found in token context"})
		return
	}
	authUserRole, roleOk := authUserRoleValue.(string)
	if !roleOk {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User role in token context is of invalid type"})
		return
	}

	if authUserRole != "admin" && authUserID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user profile"})
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	userResponse, err := h.userUsecase.UpdateUser(c.Request.Context(), targetUserID, req)
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	targetUserID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format in URL"})
		return
	}

	authUserIDValue, exists := c.Get(AuthUserIDKey)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get authenticated user ID from context"})
		return
	}
	authUserID, ok := authUserIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authenticated user ID in context is of invalid type"})
		return
	}

	authUserRoleValue, roleExists := c.Get(AuthUserRoleKey)
	if !roleExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User role not found in token context"})
		return
	}
	authUserRole, roleOk := authUserRoleValue.(string)
	if !roleOk {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User role in token context is of invalid type"})
		return
	}

	if authUserRole != "admin" && authUserID != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this user profile"})
		return
	}

	err = h.userUsecase.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrUserNotFound.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User profile deleted successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginResponse, err := h.userUsecase.Login(c.Request.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": usecase.ErrInvalidCredentials.Error()})
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": usecase.ErrInvalidCredentials.Error()})
		case usecase.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrInvalidInput.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login user"})
		}
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {
	authUserIDValue, exists := c.Get(AuthUserIDKey)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	authUserID, ok := authUserIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID format in token is invalid"})
		return
	}

	userResponse, err := h.userUsecase.GetUserByID(c.Request.Context(), authUserID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Authenticated user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		}
		return
	}

	c.JSON(http.StatusOK, userResponse)
}
