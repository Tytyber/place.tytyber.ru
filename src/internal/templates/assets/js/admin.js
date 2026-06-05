// Chart.js configuration and visual effects
document.addEventListener('DOMContentLoaded', () => {
    // Initialize all charts
    initUserDistributionChart();
    initTrafficChart();
    initVisitsTrendChart();
    initHealthChart();

    // Add visual effects
    initHoverEffects();
    initAnimations();
});

// User Distribution Donut Chart
function initUserDistributionChart() {
    const canvas = document.getElementById('userDistributionChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const data = {
        newUsers: parseInt('{{.NewUsers}}') || 3289,
        returningUsers: parseInt('{{.ReturningUsers}}') || 5080,
        inactiveUsers: parseInt('{{.InactiveUsers}}') || 1500
    };

    const total = data.newUsers + data.returningUsers + data.inactiveUsers;
    const colors = ['#ccff00', '#ffffff', '#333333'];
    const values = [data.newUsers, data.returningUsers, data.inactiveUsers];

    let startAngle = 0;
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(centerX, centerY) - 10;
    const innerRadius = radius * 0.7;

    // Draw donut
    values.forEach((value, index) => {
        const sliceAngle = (value / total) * 2 * Math.PI;

        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, startAngle, startAngle + sliceAngle);
        ctx.arc(centerX, centerY, innerRadius, startAngle + sliceAngle, startAngle, true);
        ctx.closePath();

        ctx.fillStyle = colors[index];
        ctx.fill();

        startAngle += sliceAngle;
    });
}

// Traffic Activity Chart
function initTrafficChart() {
    const canvas = document.getElementById('trafficChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const data = [4200, 3800, 5100, 4600, 5289, 4100, 3900];
    const max = Math.max(...data);

    canvas.width = canvas.offsetWidth;
    canvas.height = 200;

    const width = canvas.width;
    const height = canvas.height;
    const padding = 20;
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;

    // Draw grid
    ctx.strokeStyle = '#2a2a2a';
    ctx.lineWidth = 1;
    for (let i = 0; i <= 4; i++) {
        const y = padding + (chartHeight / 4) * i;
        ctx.beginPath();
        ctx.moveTo(padding, y);
        ctx.lineTo(width - padding, y);
        ctx.stroke();
    }

    // Draw area
    const gradient = ctx.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, 'rgba(204, 255, 0, 0.3)');
    gradient.addColorStop(1, 'rgba(204, 255, 0, 0)');

    ctx.beginPath();
    ctx.moveTo(padding, height - padding);

    data.forEach((value, index) => {
        const x = padding + (chartWidth / (data.length - 1)) * index;
        const y = height - padding - (value / max) * chartHeight;

        if (index === 0) {
            ctx.lineTo(x, y);
        } else {
            const prevX = padding + (chartWidth / (data.length - 1)) * (index - 1);
            const prevY = height - padding - (data[index - 1] / max) * chartHeight;
            const cpX = (prevX + x) / 2;
            ctx.bezierCurveTo(cpX, prevY, cpX, y, x, y);
        }
    });

    ctx.lineTo(width - padding, height - padding);
    ctx.closePath();
    ctx.fillStyle = gradient;
    ctx.fill();

    // Draw line
    ctx.beginPath();
    data.forEach((value, index) => {
        const x = padding + (chartWidth / (data.length - 1)) * index;
        const y = height - padding - (value / max) * chartHeight;

        if (index === 0) {
            ctx.moveTo(x, y);
        } else {
            const prevX = padding + (chartWidth / (data.length - 1)) * (index - 1);
            const prevY = height - padding - (data[index - 1] / max) * chartHeight;
            const cpX = (prevX + x) / 2;
            ctx.bezierCurveTo(cpX, prevY, cpX, y, x, y);
        }
    });

    ctx.strokeStyle = '#ccff00';
    ctx.lineWidth = 2;
    ctx.stroke();

    // Draw points
    data.forEach((value, index) => {
        const x = padding + (chartWidth / (data.length - 1)) * index;
        const y = height - padding - (value / max) * chartHeight;

        ctx.beginPath();
        ctx.arc(x, y, 4, 0, Math.PI * 2);
        ctx.fillStyle = '#ccff00';
        ctx.fill();
    });
}

// Daily Visits Trend Chart
function initVisitsTrendChart() {
    const canvas = document.getElementById('visitsTrendChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const data = [680, 720, 820, 920, 780, 850, 900];
    const max = Math.max(...data);

    canvas.width = canvas.offsetWidth;
    canvas.height = 200;

    const width = canvas.width;
    const height = canvas.height;
    const padding = 20;
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;

    // Draw line
    ctx.beginPath();
    data.forEach((value, index) => {
        const x = padding + (chartWidth / (data.length - 1)) * index;
        const y = height - padding - (value / max) * chartHeight;

        if (index === 0) {
            ctx.moveTo(x, y);
        } else {
            const prevX = padding + (chartWidth / (data.length - 1)) * (index - 1);
            const prevY = height - padding - (data[index - 1] / max) * chartHeight;
            const cpX = (prevX + x) / 2;
            ctx.bezierCurveTo(cpX, prevY, cpX, y, x, y);
        }
    });

    ctx.strokeStyle = '#ffffff';
    ctx.lineWidth = 2;
    ctx.stroke();

    // Draw points and labels
    data.forEach((value, index) => {
        const x = padding + (chartWidth / (data.length - 1)) * index;
        const y = height - padding - (value / max) * chartHeight;

        // Point
        ctx.beginPath();
        ctx.arc(x, y, 4, 0, Math.PI * 2);
        ctx.fillStyle = '#ffffff';
        ctx.fill();

        // Label for max value
        if (value === max) {
            ctx.fillStyle = '#ccff00';
            ctx.font = '12px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(`$${value}`, x, y - 15);
        }
    });
}

// System Health Chart
function initHealthChart() {
    const canvas = document.getElementById('healthChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const data = [85, 88, 92, 90, 95, 93, 96, 94, 97, 95, 98, 96];
    const max = 100;

    canvas.width = canvas.offsetWidth;
    canvas.height = 100;

    const width = canvas.width;
    const height = canvas.height;
    const padding = 10;
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;
    const barWidth = (chartWidth / data.length) * 0.7;
    const spacing = chartWidth / data.length;

    // Draw bars
    data.forEach((value, index) => {
        const x = padding + spacing * index + (spacing - barWidth) / 2;
        const barHeight = (value / max) * chartHeight;
        const y = height - padding - barHeight;

        const gradient = ctx.createLinearGradient(0, y, 0, height - padding);
        gradient.addColorStop(0, '#ccff00');
        gradient.addColorStop(1, 'rgba(204, 255, 0, 0.3)');

        ctx.fillStyle = gradient;
        ctx.fillRect(x, y, barWidth, barHeight);
    });
}

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
    // Animate progress bars
    const progressBars = document.querySelectorAll('.progress-fill');
    progressBars.forEach(bar => {
        const width = bar.style.width;
        bar.style.width = '0%';
        setTimeout(() => {
            bar.style.width = width;
        }, 100);
    });

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

// Tab switching
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', function() {
        this.parentElement.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        this.classList.add('active');
    });
});

// Theme toggle
document.querySelectorAll('.theme-toggle').forEach(btn => {
    btn.addEventListener('click', function() {
        document.querySelectorAll('.theme-toggle').forEach(b => b.classList.remove('active'));
        this.classList.add('active');
    });
});