# MagicStream — System Design

## Overview

MagicStream is a full-stack web application that simulates a modern movie streaming platform with AI-powered recommendations. The system allows users to browse movies, stream content (simulated), write reviews, and receive personalized movie suggestions based on their preferences.

### Key Features
- **Movie Streaming Simulation**: Frontend-based video playback using React Player
- **AI-Powered Recommendations**: Personalized suggestions using Google Gemini AI
- **User Authentication**: JWT-based authentication with role-based access control
- **Admin Review System**: Sentiment analysis for movie rankings
- **Secure Cross-Origin Architecture**: HTTPS with proper CORS configuration

### Technology Stack

| Component | Technology | Version/Details |
|-----------|------------|-----------------|
| **Frontend** | React, TypeScript, Vite | React 19, TypeScript 5.9 |
| **Backend** | Go, Gin Framework | Go 1.25.5, Gin 1.11.0 |
| **Database** | MongoDB | MongoDB Driver v2 |
| **AI Service** | Google GenAI (Gemini) | Cloud AI Integration |
| **Authentication** | JWT | HMAC-SHA256 signing |
| **Styling** | Bootstrap, CSS | Bootstrap 5.3.8 |
| **HTTP Client** | Axios | Axios 1.13.2 |

## Architecture

### High-Level Architecture

```
┌─────────────────┐    HTTPS    ┌─────────────────┐    ┌─────────────────┐
│   React SPA     │────────────▶│   Go Backend    │◀──▶│   MongoDB       │
│   (localhost:   │             │   (localhost:   │    │   Database      │
│    5173)        │◀────────────│    8080)        │    │                 │
└─────────────────┘             └─────────────────┘    └─────────────────┘
         │                              │                        │
         │                              │                        │
         ▼                              ▼                        ▼
┌─────────────────┐             ┌─────────────────┐    ┌─────────────────┐
│   Google GenAI  │             │   JWT Tokens    │    │   Collections:  │
│   (Gemini)      │             │   (Cookies)     │    │   users, movies │
│                 │             │                 │    │   genres,       │
└─────────────────┘             └─────────────────┘    │   rankings      │
                                                       └─────────────────┘
```

### Component Breakdown

#### Frontend Architecture
- **Framework**: React 19 with TypeScript for type safety
- **Build Tool**: Vite for fast development and optimized production builds
- **Routing**: React Router DOM for client-side navigation
- **State Management**: React Context API for authentication state
- **HTTP Client**: Axios with custom hooks for authenticated requests
- **Styling**: Bootstrap 5 for responsive design
- **Video Playback**: React Player for streaming simulation

#### Backend Architecture
- **Framework**: Gin web framework for high-performance HTTP routing
- **Database**: MongoDB with official Go driver
- **Authentication**: Custom JWT middleware with token refresh
- **CORS**: Configured for secure cross-origin requests
- **TLS**: HTTPS support with local certificates
- **External APIs**: Google GenAI integration for AI features

#### Database Architecture
- **Type**: NoSQL document database (MongoDB)
- **Collections**: users, movies, genres, rankings
- **Indexing**: Optimized queries for recommendations and search
- **Connection**: Connection pooling and error handling

## Data Model

### User Collection
```json
{
  "_id": "ObjectId",
  "user_id": "string (UUID)",
  "first_name": "string",
  "last_name": "string",
  "email": "string (unique)",
  "password": "string (bcrypt hashed)",
  "role": "string (user|admin)",
  "token": "string (current access token)",
  "refresh_token": "string (current refresh token)",
  "favourite_genres": ["string"] // Array of genre names
}
```

### Movie Collection
```json
{
  "_id": "ObjectId",
  "imdb_id": "string (unique)",
  "title": "string",
  "poster_path": "string (URL)",
  "youtube_id": "string (YouTube video ID)",
  "genre": ["string"], // Array of genre names
  "admin_review": "string (optional)",
  "ranking": {
    "name": "string (e.g., 'Excellent', 'Good')",
    "value": "number (1-5, lower is better)"
  }
}
```

### Genre Collection
```json
{
  "_id": "ObjectId",
  "genre_id": "string (UUID)",
  "genre_name": "string (unique)"
}
```

### Ranking Collection
```json
{
  "_id": "ObjectId",
  "ranking_value": "number (1-5)",
  "ranking_name": "string"
}
```

## API Design

### Public Endpoints

#### GET /movies
**Description**: Retrieve all movies with basic information
**Authentication**: None
**Response**:
```json
[
  {
    "imdb_id": "tt0111161",
    "title": "The Shawshank Redemption",
    "poster_path": "/path/to/poster.jpg",
    "genre": ["Drama", "Crime"],
    "ranking": {"name": "Excellent", "value": 1}
  }
]
```

#### GET /genres
**Description**: Retrieve all available genres
**Authentication**: None
**Response**:
```json
[
  {"genre_id": "uuid", "genre_name": "Action"},
  {"genre_id": "uuid", "genre_name": "Comedy"}
]
```

#### POST /register
**Description**: Register a new user account
**Authentication**: None
**Request**:
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "securepassword",
  "favourite_genres": ["Action", "Sci-Fi"]
}
```

#### POST /login
**Description**: Authenticate user and set JWT cookies
**Authentication**: None
**Request**:
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

#### POST /refresh-token
**Description**: Refresh access token using refresh token
**Authentication**: Valid refresh token cookie
**Response**: Sets new access_token cookie

### Protected Endpoints

#### GET /movie/:imdb_id
**Description**: Get detailed information about a specific movie
**Authentication**: Required
**Response**:
```json
{
  "imdb_id": "tt0111161",
  "title": "The Shawshank Redemption",
  "poster_path": "/path/to/poster.jpg",
  "youtube_id": "youtube_video_id",
  "genre": ["Drama", "Crime"],
  "admin_review": "A masterpiece of storytelling...",
  "ranking": {"name": "Excellent", "value": 1}
}
```

#### POST /add-movie
**Description**: Add a new movie to the database (Admin only)
**Authentication**: Required (Admin role)
**Request**:
```json
{
  "imdb_id": "tt0111161",
  "title": "The Shawshank Redemption",
  "poster_path": "/path/to/poster.jpg",
  "youtube_id": "youtube_video_id",
  "genre": ["Drama", "Crime"]
}
```

#### GET /recommended-movies
**Description**: Get personalized movie recommendations
**Authentication**: Required
**Response**: Array of recommended movies based on user's favorite genres

#### PATCH /update-review/:imdb_id
**Description**: Update admin review and ranking for a movie (Admin only)
**Authentication**: Required (Admin role)
**Request**:
```json
{
  "admin_review": "An outstanding film that stands the test of time."
}
```

#### POST /logout
**Description**: Logout user and clear authentication cookies
**Authentication**: Required
**Response**: Clears authentication cookies

## Authentication & Authorization

### JWT Token Structure

#### Access Token Claims
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "user|admin",
  "exp": 1640995200,  // 1 hour expiry
  "iat": 1640991600,
  "iss": "magic-stream"
}
```

#### Refresh Token Claims
```json
{
  "user_id": "uuid",
  "exp": 1641081600,  // 24 hour expiry
  "iat": 1640991600,
  "iss": "magic-stream"
}
```

### Cookie Configuration
- **access_token**: HttpOnly, Secure, SameSite=None, Domain=localhost, Path=/
- **refresh_token**: HttpOnly, Secure, SameSite=None, Domain=localhost, Path=/

### Authentication Flow

1. **Login**:
   - Client sends email/password
   - Server validates credentials
   - Server generates access and refresh tokens
   - Server sets HttpOnly cookies
   - Server returns success response

2. **Token Refresh**:
   - Client calls /refresh-token with refresh cookie
   - Server validates refresh token
   - Server generates new token pair
   - Server updates database with new tokens
   - Server sets new cookies

3. **API Access**:
   - Client makes authenticated request
   - AuthMiddleware validates access token
   - Injects user_id and role into context
   - Protected endpoint executes

4. **Logout**:
   - Client calls /logout
   - Server clears tokens in database
   - Server expires cookies

## AI Recommendation System

### Recommendation Algorithm

1. **User Profile Analysis**:
   - Extract user's favorite genres from database
   - Query movies matching those genres using MongoDB $in operator

2. **Ranking-Based Sorting**:
   - Sort movies by ranking.value ascending (lower values = higher rankings)
   - Apply limit (default: 5 recommendations)

3. **Query Example**:
```javascript
db.movies.find({
  genre: { $in: ["Action", "Sci-Fi"] }
}).sort({
  "ranking.value": 1
}).limit(5)
```

### Admin Review Processing

1. **Review Submission**:
   - Admin submits text review for a movie
   - System constructs AI prompt with ranking categories

2. **Gemini AI Integration**:
   - Send prompt to Google GenAI API
   - AI analyzes sentiment and classifies into ranking category
   - System maps AI response to ranking value

3. **Prompt Template Example**:
```
Analyze this movie review and classify it into one of these categories:
- Excellent (Outstanding, masterpiece)
- Good (Solid, well-made)
- Average (Decent, okay)
- Poor (Disappointing, flawed)
- Terrible (Awful, should be avoided)

Review: "{admin_review}"

Respond with only the category name.
```

## Frontend Architecture

### Component Structure
```
src/
├── components/
│   ├── Layout.tsx              // Main app layout
│   ├── RequiredAuth.tsx        // Route protection
│   ├── context/
│   │   └── AuthProvider.tsx    // Authentication context
│   ├── header/
│   │   └── Header.tsx          // Navigation header
│   ├── home/
│   │   └── Home.tsx            // Landing page
│   ├── login/
│   │   └── Login.tsx           // Login form
│   ├── register/
│   │   └── Register.tsx        // Registration form
│   ├── movies/
│   │   └── Movies.tsx          // Movie listing
│   ├── movie/
│   │   └── Movie.tsx           // Individual movie page
│   ├── recommended/
│   │   └── Recommended.tsx     // Recommendations page
│   ├── review/
│   │   └── Review.tsx          // Admin review form
│   ├── stream/
│   │   └── StreamMovie.tsx     // Video player
│   └── spinner/
│       └── Spinner.tsx         // Loading component
├── hooks/
│   ├── useAuth.tsx             // Authentication hook
│   └── useAxiosPrivate.tsx     // Private API hook
├── api/
│   └── axiosConfig.ts          // HTTP client config
└── models/
    └── movies.ts               // TypeScript interfaces
```

### Custom Hooks

#### useAuth Hook
- Manages authentication state
- Provides login/logout functions
- Persists state in localStorage
- Handles token refresh logic

#### useAxiosPrivate Hook
- Creates axios instance with interceptors
- Automatically attaches credentials
- Handles 401 responses with token refresh
- Retries failed requests after refresh

### State Management

#### Authentication Context
```typescript
interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refresh: () => Promise<void>;
}

interface User {
  user_id: string;
  email: string;
  role: string;
  first_name: string;
  last_name: string;
}
```

## Security Considerations

### Authentication Security
- **Password Hashing**: bcrypt with appropriate cost factor
- **JWT Security**: HMAC-SHA256 signing with strong secrets
- **Token Expiration**: Short-lived access tokens (1 hour)
- **Secure Cookies**: HttpOnly, Secure, SameSite=None for cross-origin

### API Security
- **Input Validation**: Go validator for request payloads
- **Rate Limiting**: Not implemented (would add in production)
- **CORS Policy**: Strict origin allowance
- **HTTPS Only**: All communications encrypted

### Data Protection
- **Sensitive Data**: Passwords never logged or exposed
- **Environment Variables**: Secrets stored securely
- **Database Security**: Connection string protection
- **Role-Based Access**: Admin endpoints protected

## Performance Considerations

### Database Optimization
- **Indexing**: Compound indexes on frequently queried fields
- **Connection Pooling**: MongoDB driver connection management
- **Query Optimization**: Efficient MongoDB queries with projections

### Frontend Performance
- **Code Splitting**: Vite handles automatic code splitting
- **Lazy Loading**: Components loaded on demand
- **Asset Optimization**: Vite optimizes bundles and assets
- **Caching**: Browser caching for static assets

### Backend Performance
- **Gin Framework**: High-performance HTTP router
- **Concurrent Requests**: Go's goroutine-based concurrency
- **Memory Management**: Efficient Go memory usage
- **API Response Caching**: Not implemented (would add Redis in production)

## Error Handling

### Backend Error Handling
- **Structured Responses**: Consistent JSON error format
- **HTTP Status Codes**: Appropriate status codes (400, 401, 403, 404, 500)
- **Logging**: Structured logging with context
- **Graceful Degradation**: Services continue operating during partial failures

### Frontend Error Handling
- **User-Friendly Messages**: Clear error messages for users
- **Fallback UI**: Graceful degradation for failed requests
- **Retry Logic**: Automatic retry for transient failures
- **Loading States**: Proper loading indicators

### External Service Errors
- **AI Service Failures**: Fallback to default ranking
- **Database Connection Issues**: Connection retry logic
- **Network Timeouts**: Configurable timeout handling

## Configuration Management

### Environment Variables

#### Backend (.env)
```
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=magic_stream
ALLOWED_ORIGINS=https://localhost:5173
SECRET_KEY=your-jwt-secret-key
SECRET_REFRESH_KEY=your-refresh-secret-key
GEMINI_API_KEY=your-gemini-api-key
BASE_PROMPT_TEMPLATE=path/to/prompt/template
RECOMMENDED_MOVIE_LIMIT=5
TLS_CERT_PATH=path/to/cert.pem
TLS_KEY_PATH=path/to/key.pem
```

#### Frontend (.env)
```
VITE_API_BASE_URL=https://localhost:8080
VITE_APP_NAME=MagicStream
```

### Configuration Loading
- **godotenv**: Loads environment variables from .env file
- **Validation**: Required variables checked at startup
- **Defaults**: Sensible defaults for optional variables
- **Security**: No sensitive data in version control

## Deployment Architecture

### Development Environment
- **Local HTTPS**: Self-signed certificates for local development
- **Docker Support**: Containerized development environment
- **Hot Reload**: Vite and Gin support hot reloading
- **Database**: Local MongoDB instance

### Production Considerations
- **Containerization**: Docker images for frontend and backend
- **Orchestration**: Kubernetes for scaling and management
- **Load Balancing**: Distribute traffic across multiple instances
- **CDN**: Static asset delivery via CDN
- **Database**: Managed MongoDB (Atlas) or clustered deployment

### CI/CD Pipeline
- **Automated Testing**: Unit and integration tests
- **Build Process**: Automated builds and deployments
- **Environment Promotion**: Dev → Staging → Production
- **Rollback Strategy**: Quick rollback capabilities

## Monitoring and Observability

### Logging
- **Structured Logging**: JSON format with consistent fields
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Context**: Request IDs, user IDs, operation details
- **Centralized**: Log aggregation for analysis

### Metrics
- **Application Metrics**: Response times, error rates, throughput
- **System Metrics**: CPU, memory, disk usage
- **Business Metrics**: User registrations, movie views, recommendations served
- **AI Metrics**: Gemini API usage and success rates

### Health Checks
- **Application Health**: Database connectivity, external service status
- **Dependency Checks**: MongoDB, Gemini API availability
- **Readiness Probes**: Ensure service is ready to accept traffic
- **Liveness Probes**: Detect and restart unhealthy containers

## Backup and Recovery

### Database Backup
- **Automated Backups**: Regular MongoDB backups
- **Point-in-Time Recovery**: Ability to restore to specific point
- **Backup Verification**: Regular backup integrity checks
- **Offsite Storage**: Secure backup storage

### Disaster Recovery
- **Multi-Region**: Database replication across regions
- **Failover**: Automatic failover for high availability
- **Data Consistency**: Ensure data integrity during recovery
- **Recovery Time Objective**: Defined RTO and RPO

## Future Enhancements

### Short Term
- **Search Functionality**: Full-text search across movies
- **User Reviews**: Allow regular users to leave reviews
- **Watch History**: Track user's viewing history
- **Continue Watching**: Resume playback functionality

### Medium Term
- **Real Streaming**: Integration with actual video streaming service
- **Social Features**: User following, movie sharing
- **Advanced AI**: Content-based recommendations, collaborative filtering
- **Mobile App**: React Native mobile application

### Long Term
- **Microservices**: Break down monolithic backend
- **Global CDN**: Worldwide content delivery
- **Analytics Dashboard**: Comprehensive admin analytics
- **Machine Learning**: Advanced recommendation algorithms

## Conclusion

MagicStream demonstrates a modern, scalable web application architecture combining React, Go, MongoDB, and AI services. The system showcases best practices in authentication, security, API design, and cross-origin communication while providing a solid foundation for a movie streaming platform.

The modular architecture allows for easy scaling, the comprehensive security measures ensure user data protection, and the AI integration provides intelligent personalization features. This design serves as an excellent example of full-stack development with modern technologies.
