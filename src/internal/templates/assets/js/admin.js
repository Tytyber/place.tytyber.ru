// Visual effects
document.addEventListener('DOMContentLoaded', () => {
    initHoverEffects();
    initAnimations();
});

// Hover Effects
function initHoverEffects() {
    const cards = document.querySelectorAll('.metric-card, .chart-card');

    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transition = 'all 0.3s ease';
        });
    });
}

// Animations
function initAnimations() {
    // Animate metric values
    const metricValues = document.querySelectorAll('.metric-value');
    metricValues.forEach(value => {
        const finalValue = value.textContent;
        value.textContent = '0';

        setTimeout(() => {
            animateValue(value, 0, parseInt(finalValue.replace(/,/g, '')), 1000);
        }, 200);
    });
}

// Animate number
function animateValue(element, start, end, duration) {
    const range = end - start;
    const increment = range / (duration / 16);
    let current = start;

    const timer = setInterval(() => {
        current += increment;
        if (current >= end) {
            current = end;
            clearInterval(timer);
        }
        element.textContent = Math.floor(current).toLocaleString();
    }, 16);
}
