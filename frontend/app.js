let carrito = [];

// ==========================================
// NAVEGACIÓN ENTRE VISTAS
// ==========================================
function showView(viewName) {
    document.getElementById('view-home').classList.add('hidden');
    document.getElementById('view-restaurant').classList.add('hidden');
    document.getElementById('view-checkout').classList.add('hidden');
    document.getElementById('view-admin').classList.add('hidden');

    document.getElementById(`view-${viewName}`).classList.remove('hidden');

    if(viewName === 'home') {
        renderCatalogo();
        renderFavoritos(); // Cargamos la agregación de favoritos
    }
    if(viewName === 'admin') {
        renderReportesAdmin(); // Cargamos las agregaciones de ventas y usuarios
    }
}

// ==========================================
// RENDERIZADO DEL CATÁLOGO Y FAVORITOS
// ==========================================
async function renderFavoritos() {
    const container = document.getElementById('favoritos-container');
    container.innerHTML = '<p class="text-sm text-gray-500">Cargando favoritos...</p>';
    
    try {
        const res = await fetch('http://localhost:8080/api/reportes/mejores-restaurantes');
        if(!res.ok) throw new Error("Error en red");
        const restaurantes = await res.json();
        
        container.innerHTML = '';
        if (!restaurantes || restaurantes.length === 0) {
            container.innerHTML = '<p class="text-sm text-gray-500">No hay favoritos aún.</p>'; return;
        }

        restaurantes.forEach(rest => {
            const id = rest.id || rest._id;
            container.innerHTML += `
                <div class="min-w-[200px] bg-white p-4 rounded-lg shadow hover:shadow-lg transition cursor-pointer border-b-4 border-yellow-400" onclick="abrirRestaurante('${id}')">
                    <h4 class="font-bold text-kraken truncate">${rest.nombre}</h4>
                    <p class="text-sm text-gray-500">${rest.categoria}</p>
                    <p class="text-sm font-bold mt-1"><i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio.toFixed(1)}</p>
                </div>
            `;
        });
    } catch(e) {
        container.innerHTML = '<p class="text-xs text-red-500">Error cargando pipeline.</p>';
    }
}

async function renderCatalogo(categoria = "", buscar = "") {
    const container = document.getElementById('catalogo-container');
    container.innerHTML = '<p class="text-center text-gray-500">Cargando restaurantes...</p>'; 

    try {
        let url = `http://localhost:8080/api/restaurantes?page=1&limit=10`;
        if (categoria) url += `&categoria=${categoria}`;
        if (buscar) url += `&nombre=${buscar}`;

        const respuesta = await fetch(url);
        const restaurantes = await respuesta.json();

        container.innerHTML = ''; 
        if (!restaurantes || restaurantes.length === 0) {
            container.innerHTML = '<p class="text-center text-gray-500">No se encontraron restaurantes.</p>';
            return;
        }

        restaurantes.forEach(rest => {
            const imgUrl = rest.imagen_perfil_id ? 
                `http://localhost:8080/api/image?id=${rest.imagen_perfil_id}` : 
                `https://placehold.co/400x200?text=${encodeURIComponent(rest.nombre)}`;

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
        container.innerHTML = '<p class="text-center text-red-500">Error conectando con el backend.</p>';
    }
}

// ==========================================
// VER DETALLE DEL RESTAURANTE (Menú y Reseñas)
// ==========================================
async function abrirRestaurante(id) {
    const container = document.getElementById('restaurant-detail-container');
    container.innerHTML = '<p class="text-center text-gray-500 my-10"><i class="fa-solid fa-spinner fa-spin text-3xl"></i><br>Cargando menú y reseñas...</p>';
    showView('restaurant');

    try {
        const resRestaurante = await fetch(`http://localhost:8080/api/restaurante?id=${id}`);
        const rest = await resRestaurante.json();

        if (!rest) {
            container.innerHTML = '<p class="text-center text-red-500">Restaurante no encontrado.</p>'; return;
        }

        const resResenas = await fetch(`http://localhost:8080/api/resenas?restaurante_id=${id}`);
        let resenas = [];
        if (resResenas.ok) {
            const data = await resResenas.json();
            if(data) resenas = data;
        }

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
                        <button onclick="agregarAlCarrito('${item.nombre}', ${item.precio})" class="bg-kraken text-white px-4 py-2 rounded hover:bg-krakenAccent transition shadow"><i class="fa-solid fa-plus"></i></button>
                    </div>
                `;
            });
        } else {
            menuHTML = '<p class="text-gray-500 italic">Este restaurante aún no tiene platillos en su menú.</p>';
        }

        let resenasHTML = '';
        if (resenas.length > 0) {
            resenas.forEach(resena => {
                let estrellas = '';
                for(let i=0; i<5; i++) estrellas += i < resena.calificacion ? '<i class="fa-solid fa-star text-yellow-400"></i>' : '<i class="fa-regular fa-star text-gray-300"></i>';
                
                // Validación para evitar errores si el lookup falla y no trae nombre
                const nombreUsuario = resena.usuario && resena.usuario.nombre_completo ? resena.usuario.nombre_completo : "Usuario Anónimo";
                
                resenasHTML += `
                    <div class="bg-gray-50 p-4 rounded-lg mb-4 border border-gray-100">
                        <div class="flex justify-between items-center mb-2">
                            <span class="font-bold text-kraken"><i class="fa-solid fa-user-circle text-gray-400"></i> ${nombreUsuario}</span>
                            <span class="text-sm">${estrellas}</span>
                        </div>
                        <p class="text-gray-700">"${resena.comentario}"</p>
                    </div>
                `;
            });
        } else {
            resenasHTML = '<p class="text-gray-500 italic">Aún no hay reseñas para este restaurante. ¡Sé el primero!</p>';
        }

        container.innerHTML = `
            <div class="bg-white p-6 rounded-lg shadow-md mb-6 border-l-4 border-krakenAccent">
                <h2 class="text-3xl font-bold text-kraken mb-2">${rest.nombre}</h2>
                <p class="text-gray-600 mb-4"><i class="fa-solid fa-tag"></i> ${rest.categoria} | <i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</p>
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div class="lg:col-span-2 bg-white p-6 rounded-lg shadow-md">
                    <h3 class="text-2xl font-bold mb-4 border-b pb-2"><i class="fa-solid fa-utensils text-krakenAccent"></i> Menú</h3>
                    ${menuHTML}
                </div>
                <div class="bg-white p-6 rounded-lg shadow-md h-fit">
                    <h3 class="text-xl font-bold mb-4 border-b pb-2"><i class="fa-solid fa-comments text-krakenAccent"></i> Reseñas <span class="text-sm font-normal text-gray-500">(${resenas.length})</span></h3>
                    <div class="max-h-96 overflow-y-auto pr-2">${resenasHTML}</div>
                </div>
            </div>
        `;
    } catch (error) {
        container.innerHTML = '<p class="text-center text-red-500 mt-10">Error al cargar la información.</p>';
    }
}

// ==========================================
// RENDERIZAR AGREGACIONES EN EL ADMIN
// ==========================================
async function renderReportesAdmin() {
    const ventasDiv = document.getElementById('admin-ventas-container');
    const usuariosDiv = document.getElementById('admin-usuarios-container');
    
    ventasDiv.innerHTML = 'Cargando...';
    usuariosDiv.innerHTML = 'Cargando...';

    try {
        const resVentas = await fetch('http://localhost:8080/api/reportes/restaurantes-mas-ventas');
        const ventas = await resVentas.json();
        
        if(ventas && ventas.length > 0) {
            let html = '<ul class="list-disc pl-5">';
            ventas.forEach(v => html += `<li><b>${v.restaurante}</b>: Q${v.total_ventas.toFixed(2)} (${v.total_ordenes} órdenes)</li>`);
            ventasDiv.innerHTML = html + '</ul>';
        } else {
            ventasDiv.innerHTML = 'No hay ventas registradas.';
        }

        const resUsr = await fetch('http://localhost:8080/api/reportes/usuarios-mas-activos');
        const usuarios = await resUsr.json();

        if(usuarios && usuarios.length > 0) {
            let html = '<ul class="list-disc pl-5">';
            usuarios.forEach(u => html += `<li><b>${u.usuario}</b>: Q${u.total_gastado.toFixed(2)} (${u.total_pedidos} pedidos)</li>`);
            usuariosDiv.innerHTML = html + '</ul>';
        } else {
            usuariosDiv.innerHTML = 'No hay actividad de usuarios.';
        }
    } catch(e) {
        ventasDiv.innerHTML = '<span class="text-red-500">Error cargando reporte</span>';
        usuariosDiv.innerHTML = '<span class="text-red-500">Error cargando reporte</span>';
    }
}

// ==========================================
// CARRITO Y CHECKOUT
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
        document.getElementById('checkout-total').innerText = '0.00'; return;
    }

    carrito.forEach(item => {
        total += item.precio;
        container.innerHTML += `<div class="flex justify-between py-2 border-b"><span>${item.nombre}</span><span class="font-bold">Q ${item.precio.toFixed(2)}</span></div>`;
    });
    document.getElementById('checkout-total').innerText = total.toFixed(2);
}

function procesarTransaccion() {
    if(carrito.length === 0) { alert("El carrito está vacío"); return; }
    // A LA ESPERA DE ALEJANDRO PARA METER EL FETCH AQUÍ
    alert("🚀 ¡Transacción ACID completada! (Simulada)\n\n1. Orden guardada.\n2. Puntos sumados a tu perfil.");
    carrito = [];
    document.getElementById('cartCount').innerText = 0;
    showView('home');
}

// ==========================================
// FUNCIONES GRIDFS (Juan)
// ==========================================
async function subirFotoGridFS() {
    const input = document.getElementById('fotoInput');
    const statusText = document.getElementById('uploadStatus');
    
    if (input.files.length === 0) {
        alert("Selecciona una imagen primero."); return;
    }

    statusText.classList.remove('hidden');
    statusText.innerText = "Subiendo archivo a MongoDB Atlas...";

    const formData = new FormData();
    formData.append("imagen", input.files[0]);

    try {
        const response = await fetch('http://localhost:8080/api/upload', {
            method: 'POST',
            body: formData
        });
        
        const data = await response.json();
        if(response.ok) {
            statusText.classList.add('text-green-600');
            statusText.innerText = `✅ ¡Archivo subido exitosamente a GridFS! (ID: ${data.imagen_id})`;
            input.value = ''; 
        } else {
            throw new Error("Error en la subida");
        }
    } catch(e) {
        statusText.classList.add('text-red-600');
        statusText.innerText = "❌ Error al subir el archivo.";
    }
}

// Inicializar la app
window.onload = () => {
    showView('home');
};