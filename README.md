# JORLIST - Shopping List App

A simple, shareable shopping list app with no account required. Create a list, share the link, and collaborate in real-time.

**Live Demo:** [shopping-list-chi-blond.vercel.app](https://shopping-list-chi-blond.vercel.app/)

## Features

- **No Account Required** - Create a list instantly and start adding items
- **Easy Sharing** - Share your list via a unique link
- **Works Everywhere** - Responsive design for mobile and desktop
- **Smart Recommendations** - Get suggestions based on your shopping history
- **Drag & Drop Sorting** - Reorder items by dragging
- **Dark Mode** - Automatic theme based on system preference
- **Multilingual** - English and German support

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Vue.js 3, Vite, Vue Router, vue-i18n |
| Backend | Go REST API |
| Database | Supabase (PostgreSQL) |
| Hosting | Vercel (frontend) + Railway (backend) |

## Local Development

### Prerequisites

- Go 1.22+
- Node.js 20.19+ or 22.12+
- A Supabase project with PostgreSQL

### Backend

```bash
cd backend

# Create .env file
echo "DATABASE_URL=your_supabase_connection_string" > .env

# Run the server
go run .
# Server runs on http://localhost:8080
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run dev server
npm run dev
# App runs on http://localhost:5173
```

### Database Setup

Run the migrations in your Supabase SQL editor:

1. Create the base tables (lists, items)
2. Run `backend/migrations/001_item_history.sql` for recommendations

## API Endpoints

```
POST   /api/lists                     Create a new list
GET    /api/lists/{id}                Get list by ID
PATCH  /api/lists/{id}                Update list
DELETE /api/lists/{id}                Delete list

GET    /api/lists/{listId}/items      Get all items in a list
POST   /api/lists/{listId}/items      Add item to list
PATCH  /api/lists/{listId}/items/{id} Update item
DELETE /api/lists/{listId}/items/{id} Delete item
PUT    /api/lists/{listId}/items/reorder  Reorder items

GET    /api/lists/{listId}/recommendations  Get item suggestions
POST   /api/lists/{listId}/recommendations/{name}/dismiss  Dismiss suggestion

GET    /health                        Health check
```

## Environment Variables

### Backend

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | Supabase PostgreSQL connection string |
| `PORT` | Server port (default: 8080) |
| `CORS_ORIGIN` | Allowed frontend origin |

### Frontend

| Variable | Description |
|----------|-------------|
| `VITE_API_URL` | Backend API URL (default: http://localhost:8080) |

## Deployment

- **Frontend**: Automatically deployed to Vercel on push to `master`
- **Backend**: Automatically deployed to Railway on push to `master`

## Privacy

Lists are private by default - they can only be accessed by knowing the unique 32-character ID. There is no public list directory or search functionality.

## License

MIT
