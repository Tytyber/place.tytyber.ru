// Modal functionality
function openModal() {
    var modal = document.getElementById('editProfileModal');
    if (modal) {
        modal.style.display = 'block';
    }
}

function closeModal() {
    var modal = document.getElementById('editProfileModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// Close modal when clicking outside
window.onclick = function(event) {
    var modal = document.getElementById('editProfileModal');
    if (modal && event.target == modal) {
        closeModal();
    }
}
