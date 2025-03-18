# Quixit.us - A Resurrection of the Mixit

A creative exercise thru the exchange of audio samples.

## Features

- OAuth authentication (GitHub, Google, Discord)
- Sample pack creation and management
- Sample upload and download
- User submissions
- Time-windowed uploads
- Admin controls

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker
- PostgreSQL (via Docker)

## Quick Start

1. Clone the repository:

```bash
git clone https://github.com/yourusername/quixit.us.git
cd quixit.us
```

2. Install dependencies:

```bash
make install
```

3. Start development environment:

```bash
make dev
```

This will:

- Start PostgreSQL in Docker
- Launch the frontend dev server
- Start the backend API server

Visit `http://dev.quixit.us:3000` to view the application.

## Development Commands

- `make dev` - Start all components (frontend, backend, db)
- `make frontend` - Start frontend only
- `make backend` - Start backend only
- `make install` - Install dependencies
- `make setup-dev` - Setup dev environment
- `make build` - Production build
- `make test` - Run tests
- `make clean` - Cleanup
- `make db-up` - Start database
- `make db-down` - Stop database
- `make db-reset` - Reset database

## Project Structure

```
.
├── backend/         # Go backend API
│   ├── api/        # API handlers
│   ├── auth/       # Authentication
│   ├── db/         # Database
│   └── services/   # Business logic
├── frontend/       # Vue.js frontend
└── storage/        # Sample storage
```

## Environment Variables

Key environment variables:

```env
PORT=8080
GIN_MODE=debug
DEV_MODE=true
HOST_DOMAIN=dev.quixit.us
HOST_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=sample_exchange
```

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions! Here's how you can help:

1. Fork the repository on GitHub
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run the tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

For major changes:

- Open an issue first to discuss what you would like to change
- Make sure to update tests as appropriate
- Follow the existing code style and conventions
- Add or update documentation as needed

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.

## Code of Conduct

We are committed to providing a friendly, safe and welcoming environment for all. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).
