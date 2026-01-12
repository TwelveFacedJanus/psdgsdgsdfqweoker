// Пример использования API для фронтенда

const API_BASE = 'http://localhost:3000/api/v1';

// Функция для получения столов по категории
async function getTables(category = 'ALL') {
    try {
        const response = await fetch(`${API_BASE}/tables?category=${category}`);
        const data = await response.json();
        return data.tables;
    } catch (error) {
        console.error('Ошибка при получении столов:', error);
        return [];
    }
}

// Функция для присоединения к столу
async function joinTable(tableId) {
    try {
        const response = await fetch(`${API_BASE}/tables/${tableId}/join`, {
            method: 'POST'
        });
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Ошибка при присоединении к столу:', error);
        return null;
    }
}

// Функция для покидания стола
async function leaveTable(tableId) {
    try {
        const response = await fetch(`${API_BASE}/tables/${tableId}/leave`, {
            method: 'POST'
        });
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Ошибка при покидании стола:', error);
        return null;
    }
}

// Пример использования с кнопками категорий
document.addEventListener('DOMContentLoaded', function() {
    const categoryButtons = ['ALL', 'LOW', 'MID', 'VIP'];
    
    categoryButtons.forEach(category => {
        const button = document.getElementById(`btn-${category.toLowerCase()}`);
        if (button) {
            button.addEventListener('click', async () => {
                const tables = await getTables(category);
                displayTables(tables);
            });
        }
    });
});

// Функция для отображения столов
function displayTables(tables) {
    const container = document.getElementById('tables-container');
    if (!container) return;
    
    container.innerHTML = '';
    
    tables.forEach(table => {
        const tableElement = document.createElement('div');
        tableElement.className = 'table-card';
        tableElement.innerHTML = `
            <h3>${table.category} Table</h3>
            <p>Blinds: ${table.blinds}</p>
            <p>Buy-in: $${table.buy_in}</p>
            <p>Players: ${table.players}/${table.max_seats}</p>
            <button onclick="joinTable(${table.id})" 
                    ${table.players >= table.max_seats ? 'disabled' : ''}>
                ${table.players >= table.max_seats ? 'Full' : 'Join'}
            </button>
            <button onclick="leaveTable(${table.id})" 
                    ${table.players <= 0 ? 'disabled' : ''}>
                Leave
            </button>
        `;
        container.appendChild(tableElement);
    });
}

// Загрузить все столы при загрузке страницы
window.addEventListener('load', async () => {
    const tables = await getTables('ALL');
    displayTables(tables);
});