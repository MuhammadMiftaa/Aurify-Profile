package data

// Service names for logging
const (
	EnvService      = "ENV"
	DatabaseService = "DATABASE"
	MinioService    = "MINIO"
	GRPCService     = "GRPC"
	HTTPService     = "HTTP"
	ProfileService  = "PROFILE"
)

// Log messages
const (
	// Setup logs
	LogDBSetupSuccess    = "Database connection established"
	LogMinioSetupSuccess = "MinIO connection established"
	LogGRPCStarted       = "gRPC server started"
	LogHTTPStarted       = "HTTP server started"
	LogEnvVarMissing     = "Environment variable missing"

	// MinIO logs
	LogMinioBucketCreated = "MinIO bucket created"

	// Profile service logs
	LogGetProfileSuccess    = "Profile retrieved successfully"
	LogGetProfileFailed     = "Failed to retrieve profile"
	LogCreateProfileSuccess = "Profile created successfully"
	LogCreateProfileFailed  = "Failed to create profile"
	LogUpdateProfileSuccess = "Profile updated successfully"
	LogUpdateProfileFailed  = "Failed to update profile"
	LogUploadPhotoSuccess   = "Profile photo uploaded successfully"
	LogUploadPhotoFailed    = "Failed to upload profile photo"
	LogDeletePhotoSuccess   = "Profile photo deleted successfully"
	LogDeletePhotoFailed    = "Failed to delete profile photo"
	LogProfileNotFound      = "Profile not found"
	LogInvalidUserID        = "Invalid user ID"
	LogInvalidRequest       = "Invalid request"
)
