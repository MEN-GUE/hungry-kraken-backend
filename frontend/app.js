let carrito = [];
let carritoRestauranteID = "";
let restauranteViendoID = ""; 
const MI_USUARIO_ID = "69af05541bb5738222aa5388"; 

// 1. NAVEGACIÓN
function showView(viewName) {
    document.getElementById('view-home').classList.add('hidden');
    document.getElementById('view-restaurant').classList.add('hidden');
    document.getElementById('view-checkout').classList.add('hidden');
    document.getElementById('view-admin').classList.add('hidden');

    document.getElementById(`view-${viewName}`).classList.remove('hidden');

    if(viewName === 'home') { renderCatalogo(); renderFavoritos(); }
    if(viewName === 'admin') { renderReportesAdmin(); }
}

// 2. RENDERIZADO PRINCIPAL
async function renderFavoritos() {
    const container = document.getElementById('favoritos-container');
    try {
        const res = await fetch('http://localhost:8080/api/reportes/mejores-restaurantes');
        const restaurantes = await res.json();
        container.innerHTML = '';
        if (!restaurantes || restaurantes.length === 0) return;
        restaurantes.forEach(rest => {
            const id = rest.id || rest._id;
            container.innerHTML += `
                <div class="min-w-[200px] bg-white p-4 rounded-lg shadow hover:shadow-lg transition cursor-pointer border-b-4 border-yellow-400" onclick="abrirRestaurante('${id}')">
                    <h4 class="font-bold text-kraken truncate">${rest.nombre}</h4>
                    <p class="text-sm text-gray-500">${rest.categoria}</p>
                    <p class="text-sm font-bold mt-1"><i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio.toFixed(1)}</p>
                </div>`;
        });
    } catch(e) {
        console.error("Error al cargar favoritos", e);
    }
}

async function renderCatalogo(categoria = "", buscar = "") {
    const container = document.getElementById('catalogo-container');
    container.innerHTML = '<p class="text-center text-gray-500 w-full col-span-3 py-10"><i class="fa-solid fa-spinner fa-spin text-3xl"></i><br>Cargando información...</p>'; 

    try {
        let url = `http://localhost:8080/api/restaurantes?page=1&limit=10`;
        if (categoria) url += `&categoria=${categoria}`;
        if (buscar) url += `&nombre=${buscar}`;
        
        const res = await fetch(url);
        const restaurantes = await res.json();

        container.innerHTML = ''; 
        if (!restaurantes || restaurantes.length === 0) {
            container.innerHTML = '<p class="text-center text-gray-500 w-full col-span-3">No hay restaurantes disponibles.</p>'; 
            return;
        }

        restaurantes.forEach(rest => {
            const imgUrl = rest.imagen_perfil_id ? `http://localhost:8080/api/image?id=${rest.imagen_perfil_id}` : `https://placehold.co/400x200?text=${encodeURIComponent(rest.nombre)}`;
            const restId = rest.id || rest._id;

            container.innerHTML += `
                <div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition cursor-pointer" onclick="abrirRestaurante('${restId}')">
                    <img src="${imgUrl}" alt="${rest.nombre}" class="w-full h-48 object-cover">
                    <div class="p-4">
                        <div class="flex justify-between items-start mb-2">
                            <h3 class="text-xl font-bold text-kraken">${rest.nombre}</h3>
                            <span class="bg-gray-100 text-gray-800 text-xs font-semibold px-2.5 py-0.5 rounded border border-gray-300"><i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</span>
                        </div>
                        <span class="text-sm text-krakenAccent font-semibold">${rest.categoria}</span>
                    </div>
                </div>`;
        });
    } catch (e) { 
        container.innerHTML = '<p class="text-center text-red-500 w-full col-span-3">Error de conexión al servidor.</p>'; 
    }
}

async function renderReportesAdmin() {
    const ventasDiv = document.getElementById('admin-ventas-container');
    const usuariosDiv = document.getElementById('admin-usuarios-container');
    
    try {
        const resV = await fetch('http://localhost:8080/api/reportes/restaurantes-mas-ventas');
        const ventas = await resV.json();
        if(ventas && ventas.length > 0) {
            let html = '<ul class="list-disc pl-5 space-y-1">';
            ventas.forEach(v => html += `<li><span class="font-bold text-kraken">${v.restaurante}</span>: Q${v.total_ventas.toFixed(2)} <span class="text-gray-500 text-xs">(${v.total_ordenes} órdenes)</span></li>`);
            ventasDiv.innerHTML = html + '</ul>';
        }

        const resU = await fetch('http://localhost:8080/api/reportes/usuarios-mas-activos');
        const usuarios = await resU.json();
        if(usuarios && usuarios.length > 0) {
            let html = '<ul class="list-disc pl-5 space-y-1">';
            usuarios.forEach(u => html += `<li><span class="font-bold text-kraken">${u.usuario}</span>: Q${u.total_gastado.toFixed(2)} <span class="text-gray-500 text-xs">(${u.total_pedidos} pedidos)</span></li>`);
            usuariosDiv.innerHTML = html + '</ul>';
        }
    } catch(e) {
        console.error("Error al cargar reportes", e);
    }
}

async function abrirRestaurante(id) {
    restauranteViendoID = id; 
    const container = document.getElementById('restaurant-detail-container');
    container.innerHTML = '<p class="text-center text-gray-500 my-10"><i class="fa-solid fa-spinner fa-spin text-3xl"></i></p>';
    showView('restaurant');

    try {
        const resR = await fetch(`http://localhost:8080/api/restaurante?id=${id}`);
        const rest = await resR.json();
        const resRes = await fetch(`http://localhost:8080/api/resenas?restaurante_id=${id}`);
        const resenas = resRes.ok ? await resRes.json() : [];

        let menuHTML = '';
        if (rest.menu && rest.menu.length > 0) {
            rest.menu.forEach(item => {
                const itemId = item.item_id || "item_generico";
                menuHTML += `
                    <div class="flex justify-between items-center border-b py-4">
                        <div>
                            <h4 class="font-bold text-lg text-kraken">${item.nombre}</h4>
                            <p class="text-sm text-gray-500">${item.descripcion || ''}</p>
                            <p class="text-krakenAccent font-bold">Q ${item.precio.toFixed(2)}</p>
                        </div>
                        <button onclick="agregarAlCarrito('${id}', '${itemId}', '${item.nombre}', ${item.precio})" class="bg-kraken text-white px-4 py-2 rounded font-bold hover:bg-krakenAccent transition"><i class="fa-solid fa-cart-plus"></i></button>
                    </div>`;
            });
        } else {
            menuHTML = '<p class="text-gray-500 italic py-4">Sin platillos registrados.</p>';
        }

        let resenasHTML = '';
        if (resenas && resenas.length > 0) {
            resenas.forEach(r => {
                let estrellas = '';
                for(let i=0; i<5; i++) estrellas += i < r.calificacion ? '<i class="fa-solid fa-star text-yellow-400"></i>' : '<i class="fa-regular fa-star text-gray-300"></i>';
                const nombreUsuario = r.usuario && r.usuario.nombre_completo ? r.usuario.nombre_completo : "Usuario";
                resenasHTML += `
                    <div class="bg-gray-50 p-4 rounded-lg mb-4 border border-gray-100 shadow-sm">
                        <div class="flex justify-between mb-2"><span class="font-bold"><i class="fa-solid fa-user-circle text-gray-400"></i> ${nombreUsuario}</span><span>${estrellas}</span></div>
                        <p class="text-gray-700">"${r.comentario}"</p>
                    </div>`;
            });
        } else {
            resenasHTML = '<p class="text-gray-500 italic py-4">Sin reseñas.</p>';
        }

        container.innerHTML = `
            <div class="bg-white p-6 rounded-lg shadow-md mb-6 border-l-4 border-krakenAccent flex justify-between items-center">
                <div>
                    <h2 class="text-3xl font-bold text-kraken mb-1">${rest.nombre}</h2>
                    <p class="text-gray-600"><i class="fa-solid fa-tag"></i> ${rest.categoria} | <i class="fa-solid fa-star text-yellow-400"></i> ${rest.calificacion_promedio}</p>
                </div>
                <div class="text-right">
                    <span class="text-xs text-gray-400 block">ID Referencia:</span>
                    <span class="text-sm font-mono bg-gray-100 p-1 rounded">${rest.id || rest._id}</span>
                </div>
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div class="lg:col-span-2 bg-white p-6 rounded shadow"><h3 class="text-2xl font-bold mb-4 border-b pb-2"><i class="fa-solid fa-utensils text-krakenAccent"></i> Menú</h3>${menuHTML}</div>
                <div class="bg-white p-6 rounded shadow h-fit"><h3 class="text-xl font-bold mb-4 border-b pb-2"><i class="fa-solid fa-comments text-krakenAccent"></i> Reseñas</h3><div class="max-h-96 overflow-y-auto pr-2">${resenasHTML}</div></div>
            </div>`;
    } catch (e) { 
        container.innerHTML = '<p class="text-red-500">Error procesando los datos del restaurante.</p>'; 
    }
}

// 3. FUNCIONES DE PEDIDO Y CARRITO
function agregarAlCarrito(restId, itemId, nombre, precio) {
    if (carrito.length > 0 && carritoRestauranteID !== restId) {
        if(confirm("Solo puede procesar un pedido por restaurante. ¿Desea vaciar el carrito actual?")) {
            carrito = [];
        } else {
            return;
        }
    }
    
    carritoRestauranteID = restId;
    
    let itemExistente = carrito.find(i => i.item_id === itemId);
    if (itemExistente) {
        itemExistente.cantidad += 1;
        itemExistente.subtotal = itemExistente.cantidad * itemExistente.precio_unitario;
    } else {
        carrito.push({
            item_id: itemId,
            nombre: nombre,
            cantidad: 1,
            precio_unitario: parseFloat(precio),
            subtotal: parseFloat(precio)
        });
    }
    
    actualizarVistaCheckout();
    alert(`Se agregó ${nombre} al carrito.`);
}

function actualizarVistaCheckout() {
    const container = document.getElementById('checkout-items');
    let total = 0;
    let cantTotal = 0;
    container.innerHTML = '';

    if (carrito.length === 0) {
        container.innerHTML = '<p class="text-gray-500 italic text-center py-4">El carrito está vacío.</p>';
        document.getElementById('checkout-total').innerText = '0.00'; 
        document.getElementById('cartCount').innerText = '0';
        return;
    }

    carrito.forEach(item => {
        total += item.subtotal;
        cantTotal += item.cantidad;
        container.innerHTML += `<div class="flex justify-between py-3 border-b border-gray-100"><span class="text-gray-700">${item.cantidad}x ${item.nombre}</span><span class="font-bold text-kraken">Q ${item.subtotal.toFixed(2)}</span></div>`;
    });
    
    document.getElementById('checkout-total').innerText = total.toFixed(2);
    document.getElementById('cartCount').innerText = cantTotal;
}

async function procesarTransaccion() {
    if (carrito.length === 0) { alert("El carrito está vacío."); return; }
    
    const metodoPago = document.getElementById('metodoPagoSelect').value;
    const btn = document.getElementById('btnPagar');
    btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Procesando Pago...';
    btn.disabled = true;

    try {
        const response = await fetch('http://localhost:8080/api/checkout', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                usuario_id: MI_USUARIO_ID,
                restaurante_id: carritoRestauranteID,
                items: carrito,
                metodo_pago: metodoPago
            })
        });

        const data = await response.json();

        if (response.ok) {
            alert(`Pedido completado con éxito.\nOrden ID: ${data.orden_id}\nPuntos de lealtad obtenidos: ${data.puntos_ganados}`);
            carrito = [];
            carritoRestauranteID = "";
            actualizarVistaCheckout();
            showView('home');
        } else {
            alert("El pago no pudo ser procesado: " + data.error);
        }
    } catch (error) {
        alert("Error de conexión al intentar procesar el pago.");
    } finally {
        btn.innerHTML = '<i class="fa-solid fa-lock"></i> Confirmar Pedido';
        btn.disabled = false;
    }
}

// 4. FUNCIONES DE ESCRITURA Y MANTENIMIENTO
async function enviarResena() {
    const comentario = document.getElementById('resenaComentario').value;
    const calif = parseInt(document.getElementById('resenaCalif').value);
    if(!comentario || !calif) { alert("Completar todos los campos para la reseña."); return; }

    try {
        await fetch('http://localhost:8080/api/resenas/crear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ restaurante_id: restauranteViendoID, usuario_id: MI_USUARIO_ID, comentario: comentario, calificacion: calif })
        });
        alert("Reseña registrada.");
        document.getElementById('resenaComentario').value = '';
        document.getElementById('resenaCalif').value = '';
        abrirRestaurante(restauranteViendoID);
    } catch(e) { 
        alert("Error al registrar la reseña."); 
    }
}

async function crearRestaurante() {
    const nombre = document.getElementById('nuevoRestNombre').value;
    const cat = document.getElementById('nuevoRestCat').value;
    const lat = parseFloat(document.getElementById('nuevoRestLat').value) || 0;
    const lon = parseFloat(document.getElementById('nuevoRestLon').value) || 0;
    const fileInput = document.getElementById('nuevoRestFoto');
    const statusText = document.getElementById('crearRestStatus');
    
    if(!nombre || !cat) { alert("Completar la información principal del restaurante."); return; }

    statusText.classList.remove('hidden');
    statusText.className = "mt-2 text-sm text-blue-600 font-bold text-center";
    let imagenId = ""; 

    try {
        if (fileInput.files.length > 0) {
            statusText.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Subiendo imagen...';
            const formData = new FormData();
            formData.append("imagen", fileInput.files[0]);
            
            const response = await fetch('http://localhost:8080/api/upload', { method: 'POST', body: formData });
            if(response.ok) { 
                const data = await response.json(); 
                imagenId = data.imagen_id; 
            } else {
                throw new Error("Problema al procesar el archivo");
            }
        }

        statusText.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Guardando datos...';
        
        await fetch('http://localhost:8080/api/restaurantes/crear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ nombre: nombre, categoria: cat, imagen_perfil_id: imagenId, latitud: lat, longitud: lon })
        });

        statusText.className = "mt-2 text-sm text-green-600 font-bold text-center";
        statusText.innerText = "Restaurante registrado correctamente.";
        
        document.getElementById('nuevoRestNombre').value = '';
        document.getElementById('nuevoRestCat').value = '';
        document.getElementById('nuevoRestLat').value = '';
        document.getElementById('nuevoRestLon').value = '';
        fileInput.value = '';
        
        renderCatalogo(); 
    } catch(e) { 
        statusText.className = "mt-2 text-sm text-red-600 text-center"; 
        statusText.innerText = "Error de sistema: " + e.message; 
    }
}

async function agregarPlatillo() {
    const id = document.getElementById('menuRestId').value;
    const nombre = document.getElementById('menuNombre').value;
    const precio = parseFloat(document.getElementById('menuPrecio').value);
    const desc = document.getElementById('menuDesc').value;

    if(!id || !nombre || isNaN(precio)) { alert("Datos del platillo incompletos."); return; }

    try {
        await fetch('http://localhost:8080/api/menu/agregar', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ restaurante_id: id, nombre: nombre, precio: precio, descripcion: desc })
        });
        alert("Platillo guardado en el menú.");
        document.getElementById('menuNombre').value = '';
        document.getElementById('menuPrecio').value = '';
        document.getElementById('menuDesc').value = '';
    } catch(e) { 
        alert("Error de conectividad al agregar platillo."); 
    }
}

async function aplicarDescuentoMasivo() {
    const categoria = document.getElementById('descCategoria').value;
    if(!categoria) { alert("Ingresar la categoría a modificar."); return; }

    try {
        const response = await fetch('http://localhost:8080/api/restaurantes/descuento', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ categoria: categoria, descuento: 0.10 }) 
        });
        const data = await response.json();
        alert(`Mantenimiento aplicado.\nRegistros modificados: ${data.restaurantes_afectados}`);
        document.getElementById('descCategoria').value = '';
        renderCatalogo(); 
    } catch(e) {
        alert("Error al ejecutar proceso de descuento.");
    }
}

async function eliminarSpamMasivo() {
    if(!confirm("Esta acción eliminará todos los registros de baja puntuación. ¿Desea continuar?")) return;

    try {
        const response = await fetch('http://localhost:8080/api/resenas/masivo', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' }
        });
        const data = await response.json();
        alert(`Mantenimiento ejecutado.\nRegistros eliminados: ${data.resenas_borradas}`);
    } catch(e) {
        alert("Error al ejecutar proceso de limpieza.");
    }
}

window.onload = () => { 
    showView('home'); 
};