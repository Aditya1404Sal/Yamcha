package dashboard

const dashboardJS = `
        class YamchaDashboard {
            constructor() {
                this.ws = null;
                this.sessions = {};
                this.charts = {};
                this.realtimeChart = null;
                this.connectionStatus = document.getElementById('connectionStatus');
                
                this.initWebSocket();
                this.initForm();
                this.initRealtimeChart();
            }

            initWebSocket() {
                const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                const wsUrl = protocol + '//' + window.location.host + '/ws';
                
                this.ws = new WebSocket(wsUrl);
                
                this.ws.onopen = () => {
                    console.log('Connected to Yamcha dashboard');
                    this.updateConnectionStatus(true);
                };
                
                this.ws.onmessage = (event) => {
                    const message = JSON.parse(event.data);
                    this.handleWebSocketMessage(message);
                };
                
                this.ws.onclose = () => {
                    console.log('Disconnected from Yamcha dashboard');
                    this.updateConnectionStatus(false);
                    // Attempt to reconnect after 3 seconds
                    setTimeout(() => this.initWebSocket(), 3000);
                };
                
                this.ws.onerror = (error) => {
                    console.error('WebSocket error:', error);
                    this.updateConnectionStatus(false);
                };
            }

            updateConnectionStatus(connected) {
                if (connected) {
                    this.connectionStatus.textContent = '🟢 Connected';
                    this.connectionStatus.className = 'connection-status connected';
                } else {
                    this.connectionStatus.textContent = '🔴 Disconnected';
                    this.connectionStatus.className = 'connection-status disconnected';
                }
            }

            handleWebSocketMessage(message) {
                switch (message.type) {
                    case 'sessions_update':
                        this.updateSessionsList(message.data);
                        break;
                    case 'session_created':
                        this.addSession(message.data);
                        break;
                    case 'session_update':
                        this.updateSession(message.data);
                        break;
                    case 'session_completed':
                        this.updateSession(message.data);
                        break;
                    case 'session_deleted':
                        this.removeSession(message.data.id);
                        break;
                }
            }

            initForm() {
                const form = document.getElementById('testForm');
                form.addEventListener('submit', (e) => {
                    e.preventDefault();
                    this.createSession();
                });
            }

            initRealtimeChart() {
                const ctx = document.getElementById('realtimeChart').getContext('2d');
                this.realtimeChart = new Chart(ctx, {
                    type: 'line',
                    data: {
                        datasets: []
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    unit: 'second'
                                },
                                title: {
                                    display: true,
                                    text: 'Time'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Response Time (ms)'
                                }
                            }
                        },
                        plugins: {
                            legend: {
                                display: true,
                                position: 'top'
                            }
                        }
                    }
                });
            }

            async createSession() {
                const formData = new FormData(document.getElementById('testForm'));
                
                let headers = {};
                try {
                    if (formData.get('headers').trim()) {
                        headers = JSON.parse(formData.get('headers'));
                    }
                } catch (e) {
                    alert('Invalid JSON in headers field');
                    return;
                }

                const config = {
                    name: formData.get('testName'),
                    config: {
                        url: formData.get('url'),
                        load: {
                            attack_type: formData.get('attackType'),
                            request_count: parseInt(formData.get('requestCount')),
                            rate: parseInt(formData.get('rate')),
                            timeout: parseInt(formData.get('timeout')) + 's',
                            workers: parseInt(formData.get('workers'))
                        },
                        request: {
                            method: formData.get('method'),
                            headers: headers,
                            body: formData.get('body')
                        }
                    }
                };

                try {
                    // Create session
                    const createResponse = await fetch('/api/sessions', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify(config)
                    });
                    
                    if (!createResponse.ok) throw new Error('Failed to create session');
                    
                    const session = await createResponse.json();
                    
                    // Start session
                    const startResponse = await fetch('/api/sessions/' + session.id + '/start', {
                        method: 'POST'
                    });
                    
                    if (!startResponse.ok) throw new Error('Failed to start session');
                    
                    // Reset form
                    document.getElementById('testForm').reset();
                    
                } catch (error) {
                    alert('Error creating session: ' + error.message);
                }
            }

            updateSessionsList(sessions) {
                this.sessions = {};
                sessions.forEach(session => {
                    this.sessions[session.id] = session;
                });
                this.renderSessions();
            }

            addSession(session) {
                this.sessions[session.id] = session;
                this.renderSessions();
            }

            updateSession(session) {
                this.sessions[session.id] = session;
                this.renderSessions();
                this.updateRealtimeChart();
            }

            removeSession(sessionId) {
                delete this.sessions[sessionId];
                this.renderSessions();
                this.updateRealtimeChart();
            }

            renderSessions() {
                const container = document.getElementById('sessionsContainer');
                const noSessions = document.getElementById('noSessions');
                const chartContainer = document.getElementById('chartContainer');
                
                const sessionIds = Object.keys(this.sessions);
                
                if (sessionIds.length === 0) {
                    noSessions.style.display = 'block';
                    chartContainer.style.display = 'none';
                    return;
                }
                
                noSessions.style.display = 'none';
                chartContainer.style.display = 'block';
                
                // Remove no sessions message and render sessions grid
                const existingGrid = container.querySelector('.sessions-grid');
                if (existingGrid) existingGrid.remove();
                
                const grid = document.createElement('div');
                grid.className = 'sessions-grid';
                
                sessionIds.forEach(sessionId => {
                    const session = this.sessions[sessionId];
                    grid.appendChild(this.createSessionCard(session));
                });
                
                container.appendChild(grid);
            }

            createSessionCard(session) {
                const card = document.createElement('div');
                card.className = 'session-card';
                card.innerHTML = this.getSessionCardHTML(session);
                
                // Add event listeners for controls
                const startBtn = card.querySelector('.start-btn');
                const stopBtn = card.querySelector('.stop-btn');
                const deleteBtn = card.querySelector('.delete-btn');
                
                if (startBtn) {
                    startBtn.addEventListener('click', () => this.startSession(session.id));
                }
                if (stopBtn) {
                    stopBtn.addEventListener('click', () => this.stopSession(session.id));
                }
                if (deleteBtn) {
                    deleteBtn.addEventListener('click', () => this.deleteSession(session.id));
                }
                
                return card;
            }

            getSessionCardHTML(session) {
                const stats = session.live_stats || {};
                const canStart = session.status === 'created';
                const canStop = session.status === 'running';
                const canDelete = session.status !== 'running';
                
                let statsHTML = '';
                if (session.status === 'running' && stats.total_requests) {
                    statsHTML = '<div class="session-stats">' +
                        '<div class="stats-grid">' +
                            '<div class="stat-item">' +
                                '<div class="stat-value">' + (stats.total_requests || 0) + '</div>' +
                                '<div class="stat-label">Requests</div>' +
                            '</div>' +
                            '<div class="stat-item">' +
                                '<div class="stat-value">' + ((stats.success_rate || 0) * 100).toFixed(1) + '%</div>' +
                                '<div class="stat-label">Success</div>' +
                            '</div>' +
                            '<div class="stat-item">' +
                                '<div class="stat-value">' + (stats.p50_response_time ? stats.p50_response_time.toFixed(0) : 0) + 'ms</div>' +
                                '<div class="stat-label">P50</div>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                }
                
                let controlsHTML = '<div class="session-controls">';
                if (canStart) controlsHTML += '<button class="button start-btn">Start</button>';
                if (canStop) controlsHTML += '<button class="button danger stop-btn">Stop</button>';
                if (canDelete) controlsHTML += '<button class="button danger delete-btn">Delete</button>';
                controlsHTML += '</div>';
                
                return '<div class="session-header">' +
                        '<div class="session-name">' + session.name + '</div>' +
                        '<div class="session-status status-' + session.status + '">' + session.status + '</div>' +
                    '</div>' +
                    '<div class="session-info">' +
                        '<div class="info-item">' +
                            '<span class="info-label">URL:</span>' +
                            '<span class="info-value">' + session.config.url + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Method:</span>' +
                            '<span class="info-value">' + session.config.request.method + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Attack:</span>' +
                            '<span class="info-value">' + session.config.load.attack_type + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Requests:</span>' +
                            '<span class="info-value">' + session.config.load.request_count + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Rate:</span>' +
                            '<span class="info-value">' + session.config.load.rate + '/sec</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Started:</span>' +
                            '<span class="info-value">' + new Date(session.start_time).toLocaleTimeString() + '</span>' +
                        '</div>' +
                    '</div>' +
                    statsHTML +
                    controlsHTML;
            }

            async startSession(sessionId) {
                try {
                    const response = await fetch('/api/sessions/' + sessionId + '/start', {
                        method: 'POST'
                    });
                    if (!response.ok) throw new Error('Failed to start session');
                } catch (error) {
                    alert('Error starting session: ' + error.message);
                }
            }

            async stopSession(sessionId) {
                try {
                    const response = await fetch('/api/sessions/' + sessionId + '/stop', {
                        method: 'POST'
                    });
                    if (!response.ok) throw new Error('Failed to stop session');
                } catch (error) {
                    alert('Error stopping session: ' + error.message);
                }
            }

            async deleteSession(sessionId) {
                if (!confirm('Are you sure you want to delete this session?')) return;
                
                try {
                    const response = await fetch('/api/sessions/' + sessionId, {
                        method: 'DELETE'
                    });
                    if (!response.ok) throw new Error('Failed to delete session');
                } catch (error) {
                    alert('Error deleting session: ' + error.message);
                }
            }

            updateRealtimeChart() {
                const runningSessions = Object.values(this.sessions).filter(s => s.status === 'running');
                
                if (runningSessions.length === 0) {
                    this.realtimeChart.data.datasets = [];
                    this.realtimeChart.update();
                    return;
                }

                const colors = ['#667eea', '#f093fb', '#4facfe', '#43e97b', '#fa709a'];
                
                this.realtimeChart.data.datasets = runningSessions.map((session, index) => {
                    const stats = session.live_stats || {};
                    const color = colors[index % colors.length];
                    
                    return {
                        label: session.name,
                        data: [{
                            x: new Date(),
                            y: stats.p50_response_time || 0
                        }],
                        borderColor: color,
                        backgroundColor: color + '20',
                        tension: 0.1,
                        pointRadius: 4
                    };
                });
                
                this.realtimeChart.update();
            }
        }

        // Initialize dashboard when page loads
        document.addEventListener('DOMContentLoaded', () => {
            new YamchaDashboard();
        });
`
