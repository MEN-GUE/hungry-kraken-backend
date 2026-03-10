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
// 3. RENDERIZADO DEL CATÁLOGO (Home) - CONECTADO A API REAL
// ==========================================

async function renderCatalogo(categoria = "", buscar = "") {
    const container = document.getElementById('catalogo-container');
    container.innerHTML = '<p class="text-center text-gray-500">Cargando restaurantes...</p>'; 

    try {
        // Construimos la URL con los filtros de José Pablo
        let url = `http://localhost:8080/api/restaurantes?page=1&limit=10`;
        if (categoria) url += `&categoria=${categoria}`;
        if (buscar) url += `&nombre=${buscar}`;

        const respuesta = await fetch(url);
        const restaurantes = await respuesta.json();

        container.innerHTML = ''; // Limpiar el "Cargando..."

        if (!restaurantes || restaurantes.length === 0) {
            container.innerHTML = '<p class="text-center text-gray-500">No se encontraron restaurantes.</p>';
            return;
        }

        restaurantes.forEach(rest => {
            // Usamos tu endpoint de GridFS para cargar la imagen, si no hay, pone un placeholder
            const imgUrl = rest.imagen_perfil_id ? 
                `http://localhost:8080/api/image?id=${rest.imagen_perfil_id}` : 
                `https://placehold.co/400x200?text=${encodeURIComponent(rest.nombre)}`;

            // Nota: Cambié rest._id por rest.id, porque el ID de Mongo a veces viene anidado en JSON
            const restId = rest.id || rest._id;

            const card = `
                <div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition cursor-pointer" onclick="abrirRestaurante('${restId}')">
                    <img src="${imgUrl}" alt="${rest.nombre}" class="w-full h-48 object-cover">
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
    } catch (error) {
        console.error("Error al cargar el catálogo:", error);
        container.innerHTML = '<p class="text-center text-red-500">Error conectando con el backend. ¿Está encendido el servidor Go?</p>';
    }
}

// ==========================================
// 4. VER DETALLE DEL RESTAURANTE (Conectado a API)
// ==========================================

async function abrirRestaurante(id) {
    const container = document.getElementById('restaurant-detail-container');
    
    // Mostramos la vista inmediatamente con un mensaje de carga
    container.innerHTML = '<p class="text-center text-gray-500 my-10"><i class="fa-solid fa-spinner fa-spin text-3xl"></i><br>Cargando menú y reseñas...</p>';
    showView('restaurant');

    try {
        // 1. Obtener el restaurante completo (con su menú embebido)
        const resRestaurante = await fetch(`http://localhost:8080/api/restaurante?id=${id}`);
        const rest = await resRestaurante.json();

        if (!rest) {
            container.innerHTML = '<p class="text-center text-red-500">Restaurante no encontrado.</p>';
            return;
        }

        // 2. Obtener las reseñas (Usando el Lookup de José Pablo)
        const resResenas = await fetch(`http://localhost:8080/api/resenas?restaurante_id=${id}`);
        let resenas = [];
        // Si el endpoint devuelve datos, los guardamos (evitamos error si viene null)
        if (resResenas.ok) {
            const data = await resResenas.json();
            if(data) resenas = data;
        }

        // Generar HTML del Menú (Manejo de documentos embebidos)
        let menuHTML = '';
        if (rest.menu && rest.menu.length > 0) {
            rest.menu.forEach(item => {
                menuHTML += `
                    <div class="flex justify-between items-center border-b py-4">
                        <div>
                            <h4 class="font-bold text-lg">${item.nombre}</h4>
                            <p class="text-sm text-gray-500">${item.descripcion || ''}</p>
                            <p class="text-krakenAccent font-bold">Q ${item.precio.toFixed(2)}</p>
                        </div>
                        <button onclick="agregarAlCarrito('${item.nombre}', ${item.precio})" class="bg-kraken text-white px-4 py-2 rounded hover:bg-krakenAccent transition shadow">
                            <i class="fa-solid fa-plus"></i>
                        </button>
                    </div>
                `;
            });
        } else {
            menuHTML = '<p class="text-gray-500 italic">Este restaurante aún no tiene platillos en su menú.</p>';
        }

        // Generar HTML de las Reseñas (Manejo del Lookup)
        let resenasHTML = '';
        if (resenas.length > 0) {
            resenas.forEach(resena => {
                // Generamos estrellitas visuales
                let estrellas = '';
                for(let i=0; i<5; i++) {
                    estrellas += i < resena.calificacion ? '<i class="fa-solid fa-star text-yellow-400"></i>' : '<i class="fa-regular fa-star text-gray-300"></i>';
                }

                // Usamos "resena.usuario.nombre_completo" gracias al $lookup!
                resenasHTML += `
                    <div class="bg-gray-50 p-4 rounded-lg mb-4 border border-gray-100">
                        <div class="flex justify-between items-center mb-2">
                            <span class="font-bold text-kraken"><i class="fa-solid fa-user-circle text-gray-400"></i> ${resena.usuario.nombre_completo}</span>
                            <span class="text-sm">${estrellas}</span>
                        </div>
                        <p class="text-gray-700">"${resena.comentario}"</p>
                    </div>
                `;
            });
        } else {
            resenasHTML = '<p class="text-gray-500 italic">Aún no hay reseñas para este restaurante. ¡Sé el primero!</p>';
        }

        // Inyectar todo el HTML final
        container.innerHTML = `
            <div class="bg-white p-6 rounded-lg shadow-md mb-6 border-l-4 border-krakenAccent">
                <h2 class="text-3xl font-bold text-kraken mb-2">${rest.nombre}</h2>
                <p class="text-gray-600 mb-4"><i class="fa-solid fa-tag"></i> ${rest.categoria} | <i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</p>
            </div>
            
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <!-- Columna del Menú (Izquierda, más grande) -->
                <div class="lg:col-span-2 bg-white p-6 rounded-lg shadow-md">
                    <h3 class="text-2xl font-bold mb-4 border-b pb-2"><i class="fa-solid fa-utensils text-krakenAccent"></i> Menú</h3>
                    ${menuHTML}
                </div>

                <!-- Columna de Reseñas (Derecha) -->
                <div class="bg-white p-6 rounded-lg shadow-md h-fit">
                    <h3 class="text-xl font-bold mb-4 border-b pb-2">
                        <i class="fa-solid fa-comments text-krakenAccent"></i> Reseñas 
                        <span class="text-sm font-normal text-gray-500">(${resenas.length})</span>
                    </h3>
                    <div class="max-h-96 overflow-y-auto pr-2">
                        ${resenasHTML}
                    </div>
                </div>
            </div>
        `;

    } catch (error) {
        console.error("Error al cargar detalles:", error);
        container.innerHTML = '<p class="text-center text-red-500 mt-10">Error al cargar la información del restaurante.</p>';
    }
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