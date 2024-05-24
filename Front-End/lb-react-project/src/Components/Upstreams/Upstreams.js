import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Typography, Select, MenuItem, FormControl, InputLabel, Card, List, ListItem, ListItemText, Box, Grid } from '@mui/material';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const Upstreams = () => {
    const [upstreams, setUpstreams] = useState([]);
    const [selectedUpstream, setSelectedUpstream] = useState('');

    useEffect(() => {
        const fetchHealthData = () => {
            axios.get('http://localhost:9000/upstream-health')
                .then(response => {
                    setUpstreams(response.data);
                })
                .catch(error => console.error('Error fetching health data:', error));
        };

        fetchHealthData();
        const intervalId = setInterval(fetchHealthData, 5000); // Fetch data every 5 seconds

        return () => clearInterval(intervalId); // Cleanup on unmount
    }, []);

    const handleSelectUpstream = (event) => {
        setSelectedUpstream(event.target.value);
    };

    const selectedUpstreamDetails = upstreams.find(upstream => upstream.name === selectedUpstream);

    // Function to prepare data for the pie chart
    const prepareChartData = (upstream) => {
        const healthyCount = upstream.servers.filter(server => server.status && server.lastSuccess).length;
        const unhealthyCount = upstream.servers.length - healthyCount;
        return [
            { name: 'Healthy', value: healthyCount, fillColor: '#66ff33' },
            { name: 'Unhealthy', value: unhealthyCount, fillColor: '#e60000' }
        ];
    };

    return (
        <Box sx={{ width: '100%', maxWidth: 1200, margin: '0 auto', mt: 4, textAlign: 'left' }}>
            <Typography variant="h4" gutterBottom sx={{ textAlign: 'left' }}>
                Upstreams Health Status
            </Typography>
            <FormControl fullWidth sx={{ textAlign: 'left' }}>
                <InputLabel id="upstream-select-label">Select an Upstream</InputLabel>
                <Select
                    labelId="upstream-select-label"
                    value={selectedUpstream}
                    label="Select an Upstream"
                    onChange={handleSelectUpstream}
                >
                    <MenuItem value="">
                        <em>None</em>
                    </MenuItem>
                    {upstreams.map((upstream, index) => (
                        <MenuItem key={index} value={upstream.name}>{upstream.name}</MenuItem>
                    ))}
                </Select>
            </FormControl>
            {selectedUpstreamDetails && (
                <Grid container spacing={2} sx={{ mt: 2 }}>
                    <Grid item xs={12} md={6}>
                        <Card variant="outlined" sx={{ height: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
                            <Typography variant="h6" sx={{ p: 2, width: '100%', margin: '0 auto'}}>
                                Overall Health
                            </Typography>
                            <ResponsiveContainer width="100%" height={300}>
                                <PieChart>
                                    <Pie
                                        data={prepareChartData(selectedUpstreamDetails)}
                                        dataKey="value"
                                        nameKey="name"
                                        cx="50%"
                                        cy="50%"
                                        outerRadius={100}
                                        fill="#8884d8"
                                        label
                                    >
                                        {prepareChartData(selectedUpstreamDetails).map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={entry.fillColor} />
                                        ))}
                                    </Pie>
                                    <Tooltip />
                                    <Legend layout="horizontal" align="right" verticalAlign="bottom" />
                                </PieChart>
                            </ResponsiveContainer>
                        </Card>
                    </Grid>
                    <Grid item xs={12} md={6}>
                        <Card variant="outlined" sx={{ height: '100%' }}>
                            <Typography variant="h6" sx={{ p: 2 }}>
                                Server Statuses
                            </Typography>
                            <List>
                                {selectedUpstreamDetails.servers.map((server, index) => (
                                    <ListItem key={index} divider>
                                        <ListItemText primary={`URL: ${server.url}`} secondary={`Status: ${server.status ? 'Healthy' : 'Down'}, Last Check: ${new Date(server.lastCheck).toLocaleString()}, Last Success: ${server.lastSuccess ? 'Yes' : 'No'}`} />
                                    </ListItem>
                                ))}
                            </List>
                        </Card>
                    </Grid>
                </Grid>
            )}
        </Box>
    );
};

export default Upstreams;
