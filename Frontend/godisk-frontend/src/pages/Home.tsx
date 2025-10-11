import React from 'react';

const Home: React.FC = () => {
    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh]">
            <div className="bg-white rounded-2xl shadow-lg border p-10 flex flex-col items-center max-w-lg w-full">
                <h1 className="text-4xl font-extrabold text-blue-700 mb-3 text-center">Bienvenido a GoDisk 2.0</h1>
                <p className="text-lg text-gray-700 mb-6 text-center">Administra discos, particiones y archivos de forma sencilla y visual. ¡Explora, ejecuta comandos y genera reportes fácilmente!</p>
                <div className="flex gap-4 mt-2">
                    <a href="/" className="px-5 py-2 rounded-lg bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition">Ir al Terminal</a>
                    <a href="/explorer" className="px-5 py-2 rounded-lg bg-gray-100 text-blue-700 font-semibold border border-blue-700 hover:bg-blue-50 transition">Explorar Discos</a>
                </div>
            </div>
        </div>
    );
};

export default Home;