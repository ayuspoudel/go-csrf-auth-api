# Golang Auth API

> Author: Ayush Poudel

This is a simple golang api that features RSA-signed JWT access tokens, rotating refresh tokens, CSRF protection, and HTTP-only secure cookie handling; built with net/http for maximum control and minimal attack surface.

It utilized net/http package, justinas/alice for middleware pipeline (example usage of justinas can be found in alice-usage/ at root)

## Features
- RSA-signed JWT access tokens
- Rotating refresh tokens
- CSRF protection
- HTTP-only secure cookie handling
- Middleware pipeline with justinas/alice
- Simple HTML templates for testing
- SQLite database for user storage
- Bcrypt password hashing
- Clear project structure
- Comprehensive error handling
- Comments and documentation for clarity
- Easy to extend and customize
- No external web frameworks for minimal attack surface
- Uses Go's standard net/http package for HTTP handling
- Modular middleware for authentication and CSRF protection
- Simple and intuitive API design
- Ready for production use with security best practices
- Lightweight and efficient implementation
- Focus on security and best practices