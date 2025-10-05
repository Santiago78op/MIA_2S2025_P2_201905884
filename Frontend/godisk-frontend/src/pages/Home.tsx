import React from 'react';

const Home: React.FC = () => {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
            <h1 className="text-4xl font-bold text-blue-600">Welcome to GoDisk!</h1>
            <p className="mt-4 text-lg text-gray-700">Your one-stop solution for all your disk management needs.</p>
        </div>
    );
};

export default Home;