import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import MaterialUISwitch from '../Switch';
import './Metrics.css';

const Metrics = () => {
    const [isToggled, setIsToggled] = useState(false);
    const [metricsData, setMetricsData] = useState({
        rateLimitHits: {},
        totalRequests: {},
        responseTimes: [],
        upstreamConnections: {}
    });

    useEffect(() => {
        const fetchMetrics = async () => {
            try {
                const { data } = await axios.get('http://localhost:8000/metrics');
                setMetricsData(parseMetrics(data));
            } catch (error) {
                console.error('Error fetching metrics:', error);
            }
        };

        let intervalId;
        if (isToggled) {
            fetchMetrics();
            intervalId = setInterval(fetchMetrics, 500);
        }

        return () => clearInterval(intervalId);
    }, [isToggled]);

    const handleChange = () => setIsToggled(!isToggled);

    return (
        <section id="metrics">
            <FormGroup>
            <FormControlLabel
            control={<MaterialUISwitch checked={isToggled} onChange={handleChange} />}
            label={<span style={{ color: 'white' }}>Enable Metrics</span>}
        />
            {isToggled && (
                <div className="metrics-container">
                    <div className="metrics-box">
                        <h2>Rate Limit Hits</h2>
                        {Object.entries(metricsData.rateLimitHits).map(([key, value]) => (
                            <p key={key}>{`${key}: ${value}`}</p>
                        ))}
                    </div>
                    <div className="metrics-box">
                        <h2>Total Requests</h2>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart data={Object.entries(metricsData.totalRequests).map(([key, value]) => ({ name: key, Requests: value }))}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="Requests" fill="#4f7192be" />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                    <div className="metrics-box">
                        <h2>Active Connections</h2>
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart data={Object.entries(metricsData.upstreamConnections).map(([key, value]) => ({ name: key, Connections: value }))}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="Connections" fill="#4f7192be" />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                    <div className="metrics-box">
                        <h2>Response Times</h2>
                        <ResponsiveContainer width="100%" height={300}>
                            <LineChart data={metricsData.responseTimes}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="label" />
                                <YAxis />
                                <Tooltip />
                                <Line type="monotone" dataKey="count" stroke="#8884d8" activeDot={{ r: 8 }} />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            )}
        </FormGroup>
        </section>
    );
};

const parseMetrics = (data) => {
    const lines = data.split('\n');
    const metrics = {
        rateLimitHits: {},
        totalRequests: {},
        responseTimes: [],
        upstreamConnections: {}
    };

    lines.forEach(line => {
        let matches;
        if (line.includes('loadbalancer_rate_limit_hits_total')) {
            matches = line.match(/loadbalancer_rate_limit_hits_total{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.rateLimitHits[matches[1]] = parseInt(matches[2], 10);
            }
        }

        if (line.includes('loadbalancer_requests_total')) {
            matches = line.match(/loadbalancer_requests_total{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.totalRequests[matches[1]] = parseInt(matches[2], 10);
            }
        }

        if (line.includes('loadbalancer_upstream_connections')) {
            matches = line.match(/loadbalancer_upstream_connections{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.upstreamConnections[matches[1]] = parseInt(matches[2], 10);
            }
        }

        if (line.includes('loadbalancer_response_times_milliseconds_bucket')) {
            matches = line.match(/loadbalancer_response_times_milliseconds_bucket{le="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.responseTimes.push({
                    label: `<= ${matches[1]} ms`,
                    count: parseInt(matches[2], 10)
                });
            }
        }
    });
    return metrics;
};

export default Metrics;
