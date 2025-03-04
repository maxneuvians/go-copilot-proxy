# AI Proxy Frontend

A modern React application that provides a chat interface for interacting with GitHub Copilot. Built by the Canadian Digital Service as a proof of concept for AI integration.

## Features

### Chat Interface
- Real-time chat interaction with GitHub Copilot
- Markdown support for AI responses including:
  - Code blocks with syntax highlighting
  - Tables
  - Lists
  - Formatted text
- Message history with scrollable interface
- Context-aware conversations

### UI/UX
- Clean, modern interface built with Mantine UI
- Dark/Light mode toggle
- Responsive sidebar navigation
- Fixed sidebar with navigation options

### Technical Features
- Built with React + TypeScript + Vite
- Efficient message handling with conversation history
- Real-time markdown rendering
- CORS-enabled API communication
- Component-based architecture
- CSS Modules for styling
- Responsive design principles

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:5173`

## Development

### Project Structure
```
ui/
├── src/
│   ├── components/       # Reusable components
│   ├── pages/            # Page components
│   │   ├── Chat/         # Chat interface
│   │   ├── AboutChat/    # About page
│   │   └── Navbar/       # Navigation
│   ├── services/         # API services
```

### Key Components
- `Chat`: Main chat interface with message handling
- `DarkModeToggle`: Theme switching component
- `Navbar`: Navigation and layout component
- `AboutChat`: Information about the application

### API Integration
The frontend communicates with a Go backend server that proxies requests to GitHub Copilot. The `ChatApiService` handles all API communication and maintains conversation context.
