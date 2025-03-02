package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ClientHandler handles client-related API endpoints
type ClientHandler struct {
	clientManager interface {
		ListClients() []map[string]interface{}
		GetClient(string) (interface{}, error)
		UnregisterClient(string) error
	}
}

// NewClientHandler creates a new client handler
func NewClientHandler(clientManager interface{}) *ClientHandler {
	return &ClientHandler{
		clientManager: clientManager,
	}
}

// GetClients returns a list of all clients
func (h *ClientHandler) GetClients(c *gin.Context) {
	clients := h.clientManager.ListClients()
	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
	})
}

// GetClient returns a specific client
func (h *ClientHandler) GetClient(c *gin.Context) {
	clientID := c.Param("id")
	
	client, err := h.clientManager.GetClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Client not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"client": client,
	})
}

// DeleteClient removes a client
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	clientID := c.Param("id")
	
	err := h.clientManager.UnregisterClient(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Client not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Client deleted",
	})
}
