import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Line, LineChart, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import MaterialUISwitch from '../Switch';
import './Metricstcp.css';

const TCPMetrics = () => {
    const [isToggled, setIsToggled] = useState(false);
    const [metricsData, setMetricsData] = useState({
        tcpActiveConnections: {},
        tcpBytesTransferred: {},
        tcpThroughput: {},
    });

    useEffect(() => {
        const fetchMetrics = async () => {
            try {
                const { data } = await axios.get('http://localhost:8000/metrics');
                setMetricsData(parseTCPMetrics(data));
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
                    label={<span style={{ color: 'white' }}>Enable TCP Metrics</span>}
                />
                {isToggled && (
                    <div className="metrics-container">
                        <div className="metrics-box">
                            <h2>TCP Active Connections</h2>
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart data={Object.entries(metricsData.tcpActiveConnections).map(([key, value]) => ({ name: key, Connections: value }))}>
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
                            <h2>TCP Bytes Transferred</h2>
                            <ResponsiveContainer width="100%" height={300}>
                                <LineChart data={Object.entries(metricsData.tcpBytesTransferred).map(([key, value]) => ({ name: key, Bytes: value }))}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip />
                                    <Legend />
                                    <Line type="monotone" dataKey="Bytes" stroke="#82ca9d" />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                        <div className="metrics-box">
                            <h2>TCP Throughput</h2>
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart data={Object.entries(metricsData.tcpThroughput).map(([key, value]) => ({ name: key, Requests: value }))}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip />
                                    <Legend />
                                    <Bar dataKey="Requests" fill="#4f7192be" />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>
                )}
            </FormGroup>
        </section>
    );
};

const parseTCPMetrics = (data) => {
    const lines = data.split('\n');
    const metrics = {
        tcpActiveConnections: {},
        tcpBytesTransferred: {},
        tcpThroughput: {},
    };

    lines.forEach(line => {
        let matches;
        if (line.includes('tcp_loadbalancer_active_connections')) {
            matches = line.match(/tcp_loadbalancer_active_connections{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.tcpActiveConnections[matches[1]] = parseInt(matches[2], 10);
            }
        }

        if (line.includes('tcp_loadbalancer_bytes_transferred_total')) {
            matches = line.match(/tcp_loadbalancer_bytes_transferred_total{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.tcpBytesTransferred[matches[1]] = parseInt(matches[2], 10);
            }
        }

        if (line.includes('tcp_loadbalancer_throughput_total')) {
            matches = line.match(/tcp_loadbalancer_throughput_total{upstream="([^"]+)"} (\d+)/);
            if (matches) {
                metrics.tcpThroughput[matches[1]] = parseInt(matches[2], 10);
            }
        }
    });
    return metrics;
};

export default TCPMetrics;
