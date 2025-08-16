let currentRating = 0;

document.addEventListener('DOMContentLoaded', function() {
    // Gestion du système d'étoiles
    const stars = document.querySelectorAll('.star');
    const ratingText = document.getElementById('ratingText');
    const submitButton = document.getElementById('submitReview');
    const ratingInput = document.getElementById('rating');

    const ratingTexts = {
        1: { description: "Très insatisfait" },
        2: { description: "Insatisfait" },
        3: { description: "Correct" },
        4: { description: "Satisfait" },
        5: { description: "Très satisfait" }
    };

    stars.forEach(star => {
        star.addEventListener('click', function() {
            const rating = parseInt(this.dataset.rating);
            setRating(rating);
        });

        star.addEventListener('mouseenter', function() {
            const rating = parseInt(this.dataset.rating);
            highlightStars(rating);
        });
    });

    document.getElementById('starRating').addEventListener('mouseleave', function() {
        highlightStars(currentRating);
    });

    function setRating(rating) {
        currentRating = rating;
        highlightStars(rating);
        ratingInput.value = rating

        if (ratingTexts[rating]) {
            ratingText.textContent = ratingTexts[rating].description;
        }

        // Activer le bouton de soumission
        submitButton.disabled = false;
    }

    function highlightStars(rating) {
        stars.forEach((star, index) => {
            if (index < rating) {
                star.classList.add('active');
            } else {
                star.classList.remove('active');
            }
        });
    }

    // Gestion du compteur de caractères
    const commentTextarea = document.getElementById('reviewComment');
    const characterCount = document.getElementById('characterCount');

    commentTextarea.addEventListener('input', function() {
        const length = this.value.length;
        const maxLength = 250;

        characterCount.textContent = `${length}/${maxLength} caractères`;

        // Changer la couleur selon la proximité de la limite
        characterCount.classList.remove('warning', 'error');
        if (length > maxLength * 0.9) {
            characterCount.classList.add('warning');
        }
        if (length >= maxLength) {
            characterCount.classList.add('error');
        }
    });

    // Fermer le modal en cliquant sur l'overlay
    document.getElementById('reviewModal').addEventListener('click', function(e) {
        if (e.target === this) {
            closeReviewModal();
        }
    });

    // Fermer le modal avec la touche Escape
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeReviewModal();
        }
    });
});

function openReviewModal() {
    document.getElementById('reviewModal').classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeReviewModal() {
    document.getElementById('reviewModal').classList.remove('active');
    document.body.style.overflow = '';

    // Réinitialiser le formulaire après fermeture
    setTimeout(() => {
        resetForm();
    }, 300);
}

function resetForm() {
    currentRating = 0;
    document.querySelectorAll('.star').forEach(star => star.classList.remove('active'));
    document.getElementById('ratingText').textContent = 'Cliquez sur les étoiles pour noter';
    document.getElementById('reviewComment').value = '';
    document.getElementById('characterCount').textContent = '0/250 caractères';
    document.getElementById('characterCount').classList.remove('warning', 'error');
    document.getElementById('submitReview').disabled = true;

    // Réafficher le formulaire et masquer le message de succès
    document.getElementById('reviewForm').style.display = 'block';
    document.getElementById('modalFooter').style.display = 'flex';
    document.getElementById('successMessage').style.display = 'none';
}
