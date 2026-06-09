// Terminal Visual Effects & Server Communication
class Terminal {
    constructor() {
        this.input = document.getElementById('terminalInput');
        this.output = document.getElementById('terminalOutput');
        this.commandHistory = [];
        this.historyIndex = -1;
        this.apiEndpoint = '/api/terminal'; // Go backend endpoint
        this.currentUser = 'guest'; // Default user
        this.updatePrompt(); // Set initial prompt

        this.init();
    }

    updatePrompt() {
        // Update the prompt in DOM to show current user
        const promptElements = document.querySelectorAll('.prompt');
        promptElements.forEach(el => {
            el.textContent = `${this.currentUser}@ttbr:~$`;
        });
        // Also update the input placeholder for consistency
        this.input.placeholder = `${this.currentUser}@ttbr:~$`;
    }

    init() {
        this.input.addEventListener('keydown', (e) => this.handleKeyDown(e));
        this.input.focus();
        this.printWelcome();
        this.fetchCurrentUser();

        // Keep focus on terminal
        document.addEventListener('click', (e) => {
            if (e.target.closest('.terminal-section')) {
                this.input.focus();
            }
        });
    }

    fetchCurrentUser() {
        // Fetch current user from backend
        fetch('/api/current-user')
            .then(response => response.json())
            .then(data => {
                if (data.username) {
                    this.currentUser = data.username;
                    this.updatePrompt();
                }
            })
            .catch(err => {
                // If fetch fails, user is guest
                this.currentUser = 'guest';
                this.updatePrompt();
            });
    }

    printWelcome() {
        const isDarkMode = window.location.pathname === '/dark-mode';
        const welcome = [
            '╔══════════════════════════════════════════════════════════╗',
            '                                                          ',
            isDarkMode ? '   TYTYBER DARK TERMINAL v2.0.26                          ' : '   Tytyber TERMINAL v2.0.26                               ',
            '   System is booting up..                                 ',
            isDarkMode ? '   [DARK MODE ACTIVE]                                     ' : '                                                            ',
            '╚══════════════════════════════════════════════════════════╝',
            '',
            'Connecting to server...',
            ''
        ];

        welcome.forEach((line, index) => {
            setTimeout(() => {
                this.printLine(line, index < 6 ? 'info' : '');
            }, index * 100);
        });
    }

    handleKeyDown(e) {
        if (e.key === 'Enter') {
            const command = this.input.value.trim();
            if (command) {
                this.commandHistory.push(command);
                this.historyIndex = this.commandHistory.length;
                this.sendCommand(command);
            }
            this.input.value = '';
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            if (this.historyIndex > 0) {
                this.historyIndex--;
                this.input.value = this.commandHistory[this.historyIndex];
            }
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            if (this.historyIndex < this.commandHistory.length - 1) {
                this.historyIndex++;
                this.input.value = this.commandHistory[this.historyIndex];
            } else {
                this.historyIndex = this.commandHistory.length;
                this.input.value = '';
            }
        }

        setTimeout(() => this.input.focus(), 0);
    }

    async sendCommand(command) {
        // Print command
        this.printLine(`root@cybr:~$ ${command}`, 'command');

        // Show loading
        const loadingId = this.printLine('Processing...', 'info');

        try {
            // Send to Go backend
            const response = await fetch(this.apiEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    command: command,
                    timestamp: new Date().toISOString()
                })
            });

            // Remove loading
            const loadingElement = document.getElementById(loadingId);
            if (loadingElement) {
                loadingElement.remove();
            }

            if (response.ok) {
                const data = await response.json();
                this.printResponse(data);
            } else if (response.status === 403) {
                this.printLine('Доступ запрещен. Используйте терминал для включения dark mode.', 'error');
            } else {
                this.printLine(`Error: ${response.status} ${response.statusText}`, 'error');
            }
        } catch (error) {
            // Remove loading
            const loadingElement = document.getElementById(loadingId);
            if (loadingElement) {
                loadingElement.remove();
            }

            this.printLine(`Connection error: ${error.message}`, 'error');
            this.printLine('Make sure Go backend is running', 'info');
        }

        this.scrollToBottom();
    }

    printResponse(data) {
        if (data.output) {
            // Handle clear command (check both type and output)
            if (data.type === 'clear' || data.output.toLowerCase() === 'clear') {
                this.output.innerHTML = '';
                // Print empty line after clear for better UX
                this.printLine('', 'info');
                // Update prompt after clear
                this.updatePrompt();
                return;
            }
            
            // Handle different response types
            if (data.type === 'error') {
                this.printLine(data.output, 'error');
            } else if (data.type === 'success') {
                this.printLine(data.output, 'success');
                
                // Update current user from response if provided
                if (data.current_session_user) {
                    this.currentUser = data.current_session_user;
                    this.updatePrompt();
                } else {
                    // Fallback to parsing output if current_session_user not provided
                    if (data.output.toLowerCase().includes('добро пожаловать') || data.output.toLowerCase().includes('успешно зарегистрирован')) {
                        // Extract username from output if available
                        const outputLower = data.output.toLowerCase();
                        if (outputLower.includes('добро пожаловать')) {
                            const match = data.output.match(/\b([a-zA-Z0-9_]+)\b/);
                            if (match) {
                                this.currentUser = match[0];
                                this.updatePrompt();
                            }
                        } else if (outputLower.includes('успешно зарегистрирован')) {
                            const match = data.output.match(/\b([a-zA-Z0-9_]+)\b/);
                            if (match) {
                                this.currentUser = match[0];
                                this.updatePrompt();
                            }
                        }
                    } else if (data.output.toLowerCase().includes('вы вышли из системы') || data.output.toLowerCase().includes('вы вошли как гость')) {
                        // Reset to guest user
                        this.currentUser = 'guest';
                        this.updatePrompt();
                    }
                }
                
                // Check if cookie update is needed
                if (data.update_cookie === 'true') {
                    setTimeout(() => {
                        window.location.href = '/dark-mode';
                    }, 1500);
                } else if (data.update_cookie === 'false') {
                    setTimeout(() => {
                        window.location.href = '/';
                    }, 1500);
                }
                
                // Check if redirect is needed
                if (data.redirect) {
                    setTimeout(() => {
                        window.location.href = data.redirect;
                    }, 1500);
                }
            } else if (data.type === 'table') {
                this.printTable(data.output);
            } else {
                this.printLine(data.output);
            }
        }

        if (data.additional) {
            data.additional.forEach(line => {
                this.printLine(line);
            });
        }
        
        // If waiting for input, keep focus on input and show prompt
        if (data.wait_for_input) {
            setTimeout(() => {
                this.input.focus();
            }, 100);
        }
    }

    printLine(text, className = '') {
        const line = document.createElement('div');
        const id = 'line-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        line.id = id;
        line.className = `terminal-line ${className}`;
        line.innerHTML = this.formatText(text);
        this.output.appendChild(line);
        this.scrollToBottom();
        return id;
    }

    printTable(rows) {
        rows.forEach(row => {
            const line = document.createElement('div');
            line.className = 'terminal-line';
            line.innerHTML = this.formatText(row);
            this.output.appendChild(line);
        });
        this.scrollToBottom();
    }

    formatText(text) {
        // Basic formatting for terminal output
        return text
            .replace(/\n/g, '<br>')
            .replace(/\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;');
    }

    scrollToBottom() {
        const terminal = document.getElementById('terminal');
        terminal.scrollTop = terminal.scrollHeight;
    }
}

// Visual Effects
class VisualEffects {
    constructor() {
        this.initGlitchEffect();
        this.initScanlines();
        this.initRandomGlitches();
        this.initCursorBlink();
    }

    initGlitchEffect() {
        // Add glitch effect to titles on hover
        const titles = document.querySelectorAll('.hero-title, .section-title');

        titles.forEach(title => {
            title.addEventListener('mouseenter', () => {
                this.applyGlitch(title);
            });
        });
    }

    applyGlitch(element) {
        let iterations = 0;
        const maxIterations = 10;

        const interval = setInterval(() => {
            element.style.textShadow = `
                ${Math.random() * 10 - 5}px 0 #ccff00,
                ${Math.random() * 10 - 5}px 0 #ff00ff,
                ${Math.random() * 10 - 5}px 0 #00ffff
            `;

            element.style.transform = `
                translate(${Math.random() * 4 - 2}px, ${Math.random() * 4 - 2}px)
            `;

            iterations++;

            if (iterations >= maxIterations) {
                clearInterval(interval);
                element.style.textShadow = 'none';
                element.style.transform = 'none';
            }
        }, 50);
    }

    initScanlines() {
        // Add scanline effect
        const scanline = document.createElement('div');
        scanline.className = 'scanline';
        scanline.innerHTML = `
            <style>
                .scanline {
                    position: fixed;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background: linear-gradient(
                        to bottom,
                        transparent 50%,
                        rgba(0, 0, 0, 0.1) 50%
                    );
                    background-size: 100% 4px;
                    pointer-events: none;
                    z-index: 9999;
                    opacity: 0.3;
                }
            </style>
        `;
        document.body.appendChild(scanline);
    }

    initRandomGlitches() {
        // Random glitches across the page
        setInterval(() => {
            if (Math.random() > 0.95) { // 5% chance every 3 seconds
                const elements = document.querySelectorAll('h1, h2, h3, .btn-primary');
                const randomElement = elements[Math.floor(Math.random() * elements.length)];

                randomElement.style.animation = 'none';
                setTimeout(() => {
                    randomElement.style.animation = '';
                }, 100);
            }
        }, 3000);
    }

    initCursorBlink() {
        // Custom cursor blink effect for terminal
        const style = document.createElement('style');
        style.textContent = `
            .terminal-input {
                caret-color: #ccff00;
                animation: cursorBlink 1s step-end infinite;
            }
            
            @keyframes cursorBlink {
                0%, 100% { opacity: 1; }
                50% { opacity: 0; }
            }
        `;
        document.head.appendChild(style);
    }
}

// Smooth scroll for navigation
function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

// Form handling
function initForms() {
    // Subscribe form
    const subscribeForm = document.querySelector('.subscribe-form');
    if (subscribeForm) {
        subscribeForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = e.target.querySelector('input[type="email"]').value;

            try {
                const response = await fetch('/api/subscribe', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ email })
                });

                if (response.ok) {
                    alert('Subscribed successfully. Welcome to the resistance.');
                    e.target.reset();
                } else {
                    alert('Subscription failed. Try again.');
                }
            } catch (error) {
                alert('Connection error. Please try again.');
            }
        });
    }
}

// Initialize everything when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Initialize terminal
    const terminal = new Terminal();

    // Initialize visual effects
    const effects = new VisualEffects();

    // Initialize other features
    initSmoothScroll();
    initForms();

    // Add noise effect to background
    addNoiseEffect();
});

// Add subtle noise to background
function addNoiseEffect() {
    const canvas = document.createElement('canvas');
    canvas.style.position = 'fixed';
    canvas.style.top = '0';
    canvas.style.left = '0';
    canvas.style.width = '100%';
    canvas.style.height = '100%';
    canvas.style.pointerEvents = 'none';
    canvas.style.opacity = '0.03';
    canvas.style.zIndex = '9998';
    document.body.appendChild(canvas);

    const ctx = canvas.getContext('2d');

    function resize() {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
    }

    function drawNoise() {
        const imageData = ctx.createImageData(canvas.width, canvas.height);
        const data = imageData.data;

        for (let i = 0; i < data.length; i += 4) {
            const value = Math.random() * 255;
            data[i] = value;
            data[i + 1] = value;
            data[i + 2] = value;
            data[i + 3] = 255;
        }

        ctx.putImageData(imageData, 0, 0);
        requestAnimationFrame(drawNoise);
    }

    resize();
    window.addEventListener('resize', resize);
    drawNoise();
}

// Add hover effects to cards
document.addEventListener('DOMContentLoaded', () => {
    const cards = document.querySelectorAll('.principle-card, .status-card');

    cards.forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-5px)';
            this.style.boxShadow = '0 10px 30px rgba(204, 255, 0, 0.2)';
        });

        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
            this.style.boxShadow = 'none';
        });
    });
});