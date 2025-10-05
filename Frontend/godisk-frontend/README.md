# Godisk Frontend

## Overview

Godisk Frontend is a React application built with TypeScript and styled using Tailwind CSS. This project serves as the frontend for the Godisk platform, providing a responsive and modern user interface.

## Getting Started

To get started with the project, follow these steps:

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd godisk-frontend
   ```

2. **Install dependencies:**

   Make sure you have Node.js installed. Then run:

   ```bash
   npm install
   ```

3. **Set up environment variables:**

   Create a `.env` file in the root directory based on the `.env.example` file. Update the variables as needed for your environment.

4. **Run the development server:**

   Start the development server with:

   ```bash
   npm run dev
   ```

   Your application will be available at `http://localhost:3000`.

## Project Structure

- **src/**: Contains all the source code for the application.
  - **components/**: Contains reusable components.
  - **pages/**: Contains page components for routing.
  - **styles/**: Contains global styles and Tailwind CSS directives.
  - **utils/**: Contains utility functions.
  - **types/**: Contains TypeScript interfaces and types.
  - **main.tsx**: Entry point of the application.
  
- **public/**: Contains static assets like images.

- **.env**: Environment variables for the application.

- **package.json**: Project metadata and dependencies.

- **tsconfig.json**: TypeScript configuration.

- **vite.config.ts**: Vite configuration.

- **tailwind.config.js**: Tailwind CSS configuration.

- **postcss.config.js**: PostCSS configuration.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.