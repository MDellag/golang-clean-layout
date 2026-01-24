package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIResponse representa la respuesta de una API
type APIResponse struct {
	URL        string
	StatusCode int
	Body       json.RawMessage
	Headers    http.Header
	Duration   time.Duration
}

// APIJob representa un job que hace una petición HTTP a una API
type APIJob struct {
	url      string
	method   string
	headers  map[string]string
	body     io.Reader
	client   *http.Client
	priority int
}

// NewAPIJob crea un nuevo job para hacer requests a APIs
func NewAPIJob(url, method string, priority int) *APIJob {
	return &APIJob{
		url:      url,
		method:   method,
		headers:  make(map[string]string),
		priority: priority,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithHeaders agrega headers al request
func (j *APIJob) WithHeaders(headers map[string]string) *APIJob {
	j.headers = headers
	return j
}

// WithBody agrega un body al request
func (j *APIJob) WithBody(body io.Reader) *APIJob {
	j.body = body
	return j
}

// WithClient permite usar un cliente HTTP personalizado
func (j *APIJob) WithClient(client *http.Client) *APIJob {
	j.client = client
	return j
}

// Execute implementa la interfaz GenericJob[APIResponse]
func (j *APIJob) Execute(ctx context.Context) (APIResponse, error) {
	startTime := time.Now()

	// Crear el request
	req, err := http.NewRequestWithContext(ctx, j.method, j.url, j.body)
	if err != nil {
		return APIResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Agregar headers
	for key, value := range j.headers {
		req.Header.Set(key, value)
	}

	// Ejecutar el request
	resp, err := j.client.Do(req)
	if err != nil {
		return APIResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Leer el body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(startTime)

	return APIResponse{
		URL:        j.url,
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}

// Name implementa la interfaz GenericJob
func (j *APIJob) Name() string {
	return fmt.Sprintf("api-job-%s-%s", j.method, j.url)
}

// Priority implementa la interfaz GenericJob
func (j *APIJob) Priority() int {
	return j.priority
}

// UserResponse representa un usuario de la API JSONPlaceholder
type UserResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PostResponse representa un post de la API JSONPlaceholder
type PostResponse struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// UserAPIJob es un job específico para obtener información de usuarios
type UserAPIJob struct {
	userID   int
	client   *http.Client
	priority int
}

// NewUserAPIJob crea un job para obtener información de un usuario
func NewUserAPIJob(userID int, priority int) *UserAPIJob {
	return &UserAPIJob{
		userID:   userID,
		priority: priority,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute implementa GenericJob[UserResponse]
func (j *UserAPIJob) Execute(ctx context.Context) (UserResponse, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", j.userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return UserResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return UserResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return user, nil
}

// Name implementa GenericJob
func (j *UserAPIJob) Name() string {
	return fmt.Sprintf("user-api-job-%d", j.userID)
}

// Priority implementa GenericJob
func (j *UserAPIJob) Priority() int {
	return j.priority
}

// PostAPIJob es un job específico para obtener posts
type PostAPIJob struct {
	postID   int
	client   *http.Client
	priority int
}

// NewPostAPIJob crea un job para obtener un post
func NewPostAPIJob(postID int, priority int) *PostAPIJob {
	return &PostAPIJob{
		postID:   postID,
		priority: priority,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute implementa GenericJob[PostResponse]
func (j *PostAPIJob) Execute(ctx context.Context) (PostResponse, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", j.postID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PostResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return PostResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PostResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var post PostResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return PostResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return post, nil
}

// Name implementa GenericJob
func (j *PostAPIJob) Name() string {
	return fmt.Sprintf("post-api-job-%d", j.postID)
}

// Priority implementa GenericJob
func (j *PostAPIJob) Priority() int {
	return j.priority
}
