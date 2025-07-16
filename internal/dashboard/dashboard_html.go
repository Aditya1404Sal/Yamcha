package dashboard

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Yamcha Load Testing Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            text-align: center;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }

        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }

        .controls {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }

        .controls h3 {
            margin-bottom: 15px;
            color: #667eea;
        }

        .form-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }

        .form-group {
            display: flex;
            flex-direction: column;
        }

        .form-group label {
            margin-bottom: 5px;
            font-weight: 500;
            color: #555;
        }

        .form-group input, .form-group select, .form-group textarea {
            padding: 10px;
            border: 2px solid #e1e1e1;
            border-radius: 5px;
            font-size: 14px;
            transition: border-color 0.3s;
        }

        .form-group input:focus, .form-group select:focus, .form-group textarea:focus {
            outline: none;
            border-color: #667eea;
        }

        .button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 500;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
        }

        .button:disabled {
            background: #ccc;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }

        .button.danger {
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%);
        }

        .button.danger:hover {
            box-shadow: 0 4px 15px rgba(255, 107, 107, 0.4);
        }

        .sessions-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .session-card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .session-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 20px rgba(0,0,0,0.15);
        }

        .session-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f0f0;
        }

        .session-name {
            font-size: 1.3rem;
            font-weight: 600;
            color: #333;
        }

        .session-status {
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }

        .status-created { background: #e3f2fd; color: #1976d2; }
        .status-running { background: #e8f5e8; color: #388e3c; }
        .status-completed { background: #f3e5f5; color: #7b1fa2; }
        .status-failed { background: #ffebee; color: #d32f2f; }
        .status-cancelled { background: #fff3e0; color: #f57c00; }

        .session-info {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
            margin-bottom: 15px;
            font-size: 14px;
        }

        .info-item {
            display: flex;
            justify-content: space-between;
        }

        .info-label {
            font-weight: 500;
            color: #666;
        }

        .info-value {
            color: #333;
        }

        .session-stats {
            margin: 15px 0;
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 10px;
            margin-bottom: 15px;
        }

        .stat-item {
            text-align: center;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 5px;
        }

        .stat-value {
            font-size: 1.2rem;
            font-weight: 600;
            color: #667eea;
        }

        .stat-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
        }

        .session-controls {
            display: flex;
            gap: 10px;
            margin-top: 15px;
        }

        .session-controls .button {
            flex: 1;
            padding: 8px 16px;
            font-size: 14px;
        }

        .chart-container {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }

        .chart-wrapper {
            height: 400px;
            position: relative;
        }

        .no-sessions {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }

        .no-sessions h3 {
            margin-bottom: 10px;
            color: #999;
        }

        .connection-status {
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 10px 15px;
            border-radius: 5px;
            font-size: 14px;
            font-weight: 500;
            z-index: 1000;
        }

        .connection-status.connected {
            background: #e8f5e8;
            color: #388e3c;
            border: 2px solid #81c784;
        }

        .connection-status.disconnected {
            background: #ffebee;
            color: #d32f2f;
            border: 2px solid #ef5350;
        }

        @media (max-width: 768px) {
            .form-grid {
                grid-template-columns: 1fr;
            }
            
            .sessions-grid {
                grid-template-columns: 1fr;
            }
            
            .session-info {
                grid-template-columns: 1fr;
            }
            
            .stats-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="connection-status" id="connectionStatus">Connecting...</div>
    
    <div class="header">
        <h1>🚀 Yamcha Load Testing Dashboard</h1>
        <p>Real-time load testing with live metrics and multi-session management</p>
    </div>

    <div class="container">
        <!-- Test Creation Form -->
        <div class="controls">
            <h3>Create New Load Test</h3>
            <form id="testForm">
                <div class="form-grid">
                    <div class="form-group">
                        <label for="testName">Test Name</label>
                        <input type="text" id="testName" name="testName" required placeholder="My Load Test">
                    </div>
                    <div class="form-group">
                        <label for="url">Target URL</label>
                        <input type="url" id="url" name="url" required placeholder="https://example.com/api">
                    </div>
                    <div class="form-group">
                        <label for="method">HTTP Method</label>
                        <select id="method" name="method">
                            <option value="GET">GET</option>
                            <option value="POST">POST</option>
                            <option value="PUT">PUT</option>
                            <option value="DELETE">DELETE</option>
                            <option value="PATCH">PATCH</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="attackType">Attack Pattern</label>
                        <select id="attackType" name="attackType">
                            <option value="steady">Steady</option>
                            <option value="burst">Burst</option>
                            <option value="spike">Spike</option>
                            <option value="sustained">Sustained</option>
                            <option value="gradual">Gradual</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="requestCount">Request Count</label>
                        <input type="number" id="requestCount" name="requestCount" min="1" value="100">
                    </div>
                    <div class="form-group">
                        <label for="rate">Rate (req/sec)</label>
                        <input type="number" id="rate" name="rate" min="1" value="10">
                    </div>
                    <div class="form-group">
                        <label for="timeout">Timeout (seconds)</label>
                        <input type="number" id="timeout" name="timeout" min="1" value="30">
                    </div>
                    <div class="form-group">
                        <label for="workers">Worker Count</label>
                        <input type="number" id="workers" name="workers" min="1" value="10">
                    </div>
                </div>
                <div class="form-group">
                    <label for="headers">Headers (JSON format)</label>
                    <textarea id="headers" name="headers" rows="3" placeholder="{&quot;Content-Type&quot;: &quot;application/json&quot;, &quot;Authorization&quot;: &quot;Bearer token&quot;}"></textarea>
                </div>
                <div class="form-group">
                    <label for="body">Request Body</label>
                    <textarea id="body" name="body" rows="4" placeholder="Request body content (for POST/PUT requests)"></textarea>
                </div>
                <button type="submit" class="button">Create &amp; Start Test</button>
            </form>
        </div>

        <!-- Active Sessions -->
        <div id="sessionsContainer">
            <div class="no-sessions" id="noSessions">
                <h3>No active test sessions</h3>
                <p>Create your first load test using the form above</p>
            </div>
        </div>

        <!-- Real-time Chart -->
        <div class="chart-container" id="chartContainer" style="display: none;">
            <h3>Real-time Performance Overview</h3>
            <div class="chart-wrapper">
                <canvas id="realtimeChart"></canvas>
            </div>
        </div>
    </div>

    <script>
        class YamchaDashboard {
            constructor() {
                this.ws = null;
                this.sessions = {};
                this.charts = {};
                this.realtimeChart = null;
                this.connectionStatus = document.getElementById('connectionStatus');
                this.chartData = {}; // Store historical data for each session
                
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
                console.log('📨 Received WebSocket message:', message.type, message.data);
                switch (message.type) {
                    case 'sessions_update':
                        this.updateSessionsList(message.data);
                        break;
                    case 'session_created':
                        this.addSession(message.data);
                        break;
                    case 'session_update':
                        this.updateSession(message.data);
                        this.addChartDataPoint(message.data);
                        break;
                    case 'session_completed':
                        this.updateSession(message.data);
                        break;
                    case 'session_deleted':
                        console.log('📨 Received session_deleted message:', message.data);
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
                        animation: {
                            duration: 0 // Disable animations for real-time updates
                        },
                        interaction: {
                            intersect: false,
                            mode: 'index'
                        },
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    unit: 'second',
                                    displayFormats: {
                                        second: 'HH:mm:ss'
                                    }
                                },
                                title: {
                                    display: true,
                                    text: 'Time'
                                },
                                ticks: {
                                    source: 'auto',
                                    maxTicksLimit: 10
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
                            },
                            tooltip: {
                                mode: 'index',
                                intersect: false
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
                        target: {
                            url: formData.get('url'),
                            method: formData.get('method'),
                            headers: headers,
                            body: formData.get('body')
                        },
                        load: {
                            attack_type: formData.get('attackType'),
                            requests: parseInt(formData.get('requestCount')),
                            rate: parseInt(formData.get('rate')),
                            max_workers: parseInt(formData.get('workers'))
                        },
                        http: {
                            timeout: parseInt(formData.get('timeout')) + 's',
                            keep_alive: true,
                            max_idle_conns: 100,
                            max_idle_conns_per_host: 10,
                            idle_conn_timeout: '90s'
                        },
                        system: {
                            max_cpu: 0,
                            profiling: false
                        },
                        output: {
                            dir: './results',
                            formats: ['json']
                        },
                        reporting: {
                            real_time: true,
                            progress: true,
                            update_interval: '1s'
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
                // Clear chart data for sessions that no longer exist
                const newSessionIds = sessions.map(s => s.id);
                Object.keys(this.chartData).forEach(sessionId => {
                    if (!newSessionIds.includes(sessionId)) {
                        delete this.chartData[sessionId];
                    }
                });
                
                sessions.forEach(session => {
                    this.sessions[session.id] = session;
                    // Initialize chart data if it doesn't exist
                    if (!this.chartData[session.id]) {
                        this.chartData[session.id] = [];
                    }
                });
                this.renderSessions();
            }

            addSession(session) {
                this.sessions[session.id] = session;
                this.chartData[session.id] = []; // Initialize chart data for new session
                this.renderSessions();
            }

            updateSession(session) {
                this.sessions[session.id] = session;
                this.renderSessions();
            }

            removeSession(sessionId) {
                console.log('🗑️ Removing session from frontend:', sessionId);
                delete this.sessions[sessionId];
                delete this.chartData[sessionId]; // Clean up chart data
                console.log('🗑️ Cleaned up chart data for session:', sessionId);
                this.renderSessions();
                this.updateRealtimeChart();
                console.log('✅ Session removal completed:', sessionId);
            }

            addChartDataPoint(session) {
                if (!session.live_stats) return;
                
                const now = new Date();
                const responseTime = session.live_stats.median_response_time || 0;
                
                // Initialize array if it doesn't exist
                if (!this.chartData[session.id]) {
                    this.chartData[session.id] = [];
                }
                
                // Add new data point
                this.chartData[session.id].push({
                    x: now,
                    y: responseTime / 1000000 // Convert nanoseconds to milliseconds
                });
                
                // Keep only last 60 data points (about 1 minute of data)
                if (this.chartData[session.id].length > 60) {
                    this.chartData[session.id].shift();
                }
                
                console.log('📊 Added chart data point for ' + session.name + ': ' + (responseTime/1000000).toFixed(2) + 'ms at ' + now.toLocaleTimeString());
                
                // Update the chart
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
                card.setAttribute('data-session-id', session.id);
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
                                '<div class="stat-value">' + (stats.median_response_time ? (stats.median_response_time / 1000000).toFixed(0) : 0) + 'ms</div>' +
                                '<div class="stat-label">Median</div>' +
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
                            '<span class="info-value">' + session.config.target.url + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Method:</span>' +
                            '<span class="info-value">' + session.config.target.method + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Attack:</span>' +
                            '<span class="info-value">' + session.config.load.attack_type + '</span>' +
                        '</div>' +
                        '<div class="info-item">' +
                            '<span class="info-label">Requests:</span>' +
                            '<span class="info-value">' + session.config.load.requests + '</span>' +
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
                
                // Find the delete button and disable it to prevent double-clicks
                const sessionCard = document.querySelector('[data-session-id="' + sessionId + '"]');
                const deleteBtn = sessionCard ? sessionCard.querySelector('.delete-btn') : null;
                if (deleteBtn) {
                    deleteBtn.disabled = true;
                    deleteBtn.textContent = 'Deleting...';
                }
                
                try {
                    console.log('🗑️ Attempting to delete session:', sessionId);
                    const response = await fetch('/api/sessions/' + sessionId, {
                        method: 'DELETE'
                    });
                    
                    console.log('🗑️ Delete response status:', response.status);
                    
                    if (!response.ok) {
                        const errorText = await response.text();
                        // Check if it's "already deleted" vs actual error
                        if (response.status === 404 && (errorText.includes('Session not found') || errorText.includes('already been deleted'))) {
                            console.log('⚠️ Session was already deleted');
                            // Don't show error for already deleted sessions
                            return;
                        }
                        throw new Error('HTTP ' + response.status + ': ' + errorText);
                    }
                    
                    const result = await response.json();
                    console.log('✅ Session deleted successfully:', result);
                    
                } catch (error) {
                    console.error('❌ Delete session error:', error);
                    alert('Error deleting session: ' + error.message);
                    
                    // Re-enable button on error
                    if (deleteBtn) {
                        deleteBtn.disabled = false;
                        deleteBtn.textContent = 'Delete';
                    }
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
                    const color = colors[index % colors.length];
                    const data = this.chartData[session.id] || [];
                    
                    return {
                        label: session.name + ' (Median Response Time)',
                        data: data,
                        borderColor: color,
                        backgroundColor: color + '20',
                        tension: 0.1,
                        pointRadius: 2,
                        pointHoverRadius: 4,
                        fill: false
                    };
                });
                
                this.realtimeChart.update('none'); // Use 'none' animation mode for real-time updates
            }
        }

        // Initialize dashboard when page loads
        document.addEventListener('DOMContentLoaded', () => {
            new YamchaDashboard();
        });
    </script>
</body>
</html>`
