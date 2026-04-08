// showToast displays a temporary notification message.
function showToast(message, type) {
    let container = document.querySelector('.toast-container');
    if (!container) {
        container = document.createElement('div');
        container.className = 'toast-container';
        document.body.appendChild(container);
    }

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        toast.style.transition = 'all 0.3s';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// apiAction sends a POST request to the given endpoint and handles the UI response.
async function apiAction(endpoint, id, cardElement, successMessage) {
    if (!confirm('Are you sure? This action cannot be undone.')) {
        return;
    }

    const btn = cardElement.querySelector('.btn-danger');
    const originalText = btn.textContent;
    btn.disabled = true;
    btn.textContent = 'Working...';

    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: id }),
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error || 'Request failed');
        }

        showToast(successMessage, 'success');
        cardElement.style.opacity = '0';
        cardElement.style.transform = 'scale(0.95)';
        cardElement.style.transition = 'all 0.3s';
        setTimeout(() => cardElement.remove(), 300);
    } catch (err) {
        showToast(`Error: ${err.message}`, 'error');
        btn.disabled = false;
        btn.textContent = originalText;
    }
}

function unsubscribe(id, el) {
    apiAction('/api/unsubscribe', id, el.closest('.item-card'), 'Unsubscribed successfully');
}

function deletePlaylist(id, el) {
    apiAction('/api/delete-playlist', id, el.closest('.item-card'), 'Playlist deleted');
}

function deleteVideo(id, el) {
    apiAction('/api/delete-video', id, el.closest('.item-card'), 'Video deleted');
}

function removePlaylistItem(id, el) {
    apiAction('/api/remove-playlist-item', id, el.closest('.item-card'), 'Item removed from playlist');
}
