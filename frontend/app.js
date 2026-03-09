// ==========================================
// 1. MOCK DATA (Datos falsos simulando la BD)
// ==========================================

const mockRestaurantes = [
    {
        _id: "648c1a1",
        nombre: "Hungry Kraken Seafood",
        categoria: "Mariscos",
        calificacion_promedio: 4.8,
        imagen_id: "dummy1", // Simulando un ID de GridFS
        menu: [
            { item_id: "m1", nombre: "Ceviche Kraken", precio: 120.00, descripcion: "Con pulpo fresco" },
            { item_id: "m2", nombre: "Tacos de Pescado", precio: 45.00, descripcion: "Estilo Ensenada" }
        ]
    },
    {
        _id: "648c1a2",
        nombre: "Burger Leviatán",
        categoria: "Comida Rápida",
        calificacion_promedio: 4.5,
        imagen_id: "dummy2",
        menu: [
            { item_id: "m3", nombre: "Hamburguesa Doble", precio: 65.00, descripcion: "Queso cheddar y tocino" }
        ]
    }
];

let carrito = [];

// ==========================================
// 2. NAVEGACIÓN ENTRE VISTAS
// ==========================================

function showView(viewName) {
    // Ocultar todas
    document.getElementById('view-home').classList.add('hidden');
    document.getElementById('view-restaurant').classList.add('hidden');
    document.getElementById('view-checkout').classList.add('hidden');
    document.getElementById('view-admin').classList.add('hidden');

    // Mostrar la seleccionada
    document.getElementById(`view-${viewName}`).classList.remove('hidden');

    // Si vamos al home, renderizamos el catálogo
    if(viewName === 'home') renderCatalogo();
}

// ==========================================
// 3. RENDERIZADO DEL CATÁLOGO (Home)
// ==========================================

function renderCatalogo() {
    const container = document.getElementById('catalogo-container');
    container.innerHTML = ''; // Limpiar

    mockRestaurantes.forEach(rest => {
        // En el futuro, cambiaremos 'https://placehold.co/400x200' por la URL real de tu API de GridFS
        // ej: `http://localhost:8080/api/image?id=${rest.imagen_id}`
        
        const card = `
            <div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition cursor-pointer" onclick="abrirRestaurante('${rest._id}')">
                <img src="https://placehold.co/400x200?text=${encodeURIComponent(rest.nombre)}" alt="${rest.nombre}" class="w-full h-48 object-cover">
                <div class="p-4">
                    <div class="flex justify-between items-start mb-2">
                        <h3 class="text-xl font-bold text-kraken">${rest.nombre}</h3>
                        <span class="bg-gray-100 text-gray-800 text-xs font-semibold px-2.5 py-0.5 rounded border border-gray-300"><i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</span>
                    </div>
                    <span class="text-sm text-krakenAccent font-semibold">${rest.categoria}</span>
                </div>
            </div>
        `;
        container.innerHTML += card;
    });
}

// ==========================================
// 4. VER DETALLE DEL RESTAURANTE
// ==========================================

function abrirRestaurante(id) {
    const rest = mockRestaurantes.find(r => r._id === id);
    if(!rest) return;

    const container = document.getElementById('restaurant-detail-container');
    
    // Generar HTML del Menú
    let menuHTML = '';
    rest.menu.forEach(item => {
        menuHTML += `
            <div class="flex justify-between items-center border-b py-4">
                <div>
                    <h4 class="font-bold text-lg">${item.nombre}</h4>
                    <p class="text-sm text-gray-500">${item.descripcion}</p>
                    <p class="text-krakenAccent font-bold">Q ${item.precio.toFixed(2)}</p>
                </div>
                <button onclick="agregarAlCarrito('${item.nombre}', ${item.precio})" class="bg-kraken text-white px-4 py-2 rounded hover:bg-krakenAccent transition"><i class="fa-solid fa-plus"></i></button>
            </div>
        `;
    });

    container.innerHTML = `
        <div class="bg-white p-6 rounded-lg shadow-md mb-6">
            <h2 class="text-3xl font-bold text-kraken mb-2">${rest.nombre}</h2>
            <p class="text-gray-600 mb-4"><i class="fa-solid fa-tag"></i> ${rest.categoria} | <i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</p>
        </div>
        <div class="bg-white p-6 rounded-lg shadow-md">
            <h3 class="text-2xl font-bold mb-4 border-b pb-2">Menú</h3>
            ${menuHTML}
        </div>
    `;

    showView('restaurant');
}

// ==========================================
// 5. CARRITO Y CHECKOUT
// ==========================================

function agregarAlCarrito(nombre, precio) {
    carrito.push({ nombre, precio });
    document.getElementById('cartCount').innerText = carrito.length;
    actualizarVistaCheckout();
    alert(`¡${nombre} agregado al carrito!`);
}

function actualizarVistaCheckout() {
    const container = document.getElementById('checkout-items');
    let total = 0;
    container.innerHTML = '';

    if (carrito.length === 0) {
        container.innerHTML = '<p class="text-gray-500">Tu carrito está vacío.</p>';
        document.getElementById('checkout-total').innerText = '0.00';
        return;
    }

    carrito.forEach(item => {
        total += item.precio;
        container.innerHTML += `
            <div class="flex justify-between py-2 border-b">
                <span>${item.nombre}</span>
                <span class="font-bold">Q ${item.precio.toFixed(2)}</span>
            </div>
        `;
    });
    
    document.getElementById('checkout-total').innerText = total.toFixed(2);
}

function procesarTransaccion() {
    if(carrito.length === 0) {
        alert("El carrito está vacío");
        return;
    }
    // Aquí es donde el Ing. de Consistencia meterá su Endpoint POST a Go.
    alert("🚀 ¡Transacción ACID completada!\n\n1. Orden guardada.\n2. Puntos sumados a tu perfil.");
    carrito = [];
    document.getElementById('cartCount').innerText = 0;
    showView('home');
}

// ==========================================
// 6. FUNCIONES DE TU PARTE (GRIDFS)
// ==========================================

function subirFotoGridFS() {
    const input = document.getElementById('fotoInput');
    const statusText = document.getElementById('uploadStatus');
    
    if (input.files.length === 0) {
        alert("Selecciona una imagen primero.");
        return;
    }

    statusText.classList.remove('hidden');
    statusText.innerText = "Subiendo archivo a MongoDB Atlas...";

    // Aquí conectarás con tu endpoint real POST /api/upload
    // Por ahora lo simulamos con un setTimeout
    setTimeout(() => {
        statusText.classList.add('text-green-600');
        statusText.innerText = "✅ ¡Archivo subido exitosamente a GridFS! (ID: 648c1a...)";
        input.value = ''; // Limpiar input
    }, 1500);
}

// Inicializar la app
window.onload = () => {
    showView('home');
};