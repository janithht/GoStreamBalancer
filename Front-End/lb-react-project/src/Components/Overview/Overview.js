import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { PieChart, Pie, Cell, ResponsiveContainer, LineChart, Line, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';
import { Typography, Box, Grid } from '@mui/material';
import './Overview.css';

function Overview() {
    const [serverHealth, setServerHealth] = useState({ status: 'Loading...', color: '#ccc' });
    const [healthHistory, setHealthHistory] = useState([]);

    const fetchHealthStatus = async () => {
        try {
            const response = await axios.get('http://localhost:9000/healthCheck');
            const isHealthy = response.status === 200;
            const newStatus = {
                status: isHealthy ? 'Healthy' : 'Unhealthy',
                color: isHealthy ? '#4f7192be' : '#FF6347',
                timestamp: new Date().toLocaleTimeString()
            };
            setServerHealth(newStatus);
            setHealthHistory(prevHistory => [...prevHistory, newStatus]);
        } catch (err) {
            const errorStatus = { status: 'Unhealthy', color: '#FF6347', timestamp: new Date().toLocaleTimeString() };
            setServerHealth(errorStatus);
            setHealthHistory(prevHistory => [...prevHistory, errorStatus]);
        }
    };

    useEffect(() => {
        const intervalId = setInterval(fetchHealthStatus, 1000); // Poll every second
        return () => clearInterval(intervalId);
    }, []);

    const pieData = [{ name: serverHealth.status, value: 1, fill: serverHealth.color }];

    return (
        <section id="overview">
            <div className="overview">
                <Grid container spacing={4}>
                    <Grid item xs={12} md={9}>
                        <Box className="chartBox">
                            <ResponsiveContainer width="40%" height={200}>
                                <PieChart>
                                    <Pie
                                        data={pieData}
                                        dataKey="value"
                                        innerRadius="70%"
                                        outerRadius="90%"
                                        startAngle={90}
                                        endAngle={-270}
                                        paddingAngle={2}
                                    >
                                        {pieData.map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={entry.fill} />
                                        ))}
                                    </Pie>
                                    <text
                                        x="50%"
                                        y="50%"
                                        fill="#2c3e50"
                                        textAnchor="middle"
                                        dominantBaseline="central"
                                        style={{ fontSize: '1.5em', fontWeight: 'bold', fontFamily: 'Arial, sans-serif' }}
                                    >
                                        {serverHealth.status}
                                    </text>
                                </PieChart>
                            </ResponsiveContainer>
                            <ResponsiveContainer width="60%" height={200}>
                                <LineChart data={healthHistory}>
                                    <XAxis dataKey="timestamp" />
                                    <YAxis hide />
                                    <Tooltip />
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <Line type="monotone" dataKey="status" stroke="#00BFFF" strokeWidth={2} />
                                </LineChart>
                            </ResponsiveContainer>
                        </Box>
                    </Grid>
                    <Grid item xs={12} md={3}>
                        <Box className="entryPoints">
                            <Typography variant="h6" className="entryPointsTitle">Entry Points</Typography>
                            <Typography variant="h3" className="entryPointPort">:9000</Typography>
                        </Box>
                    </Grid>
                </Grid>
            </div>
        </section>
    );
}

export default Overview;
