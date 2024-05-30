import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Container, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, TextField, Button } from '@mui/material';
import './Metricstcp.css'; // Import the CSS file

function Metricstcp() {
    const [connections, setConnections] = useState([]);
    const [loading, setLoading] = useState(true);
    const [filters, setFilters] = useState({
        client_ip: '',
        server_url: '',
        start_date: '',
        end_date: ''
    });

    const fetchConnections = () => {
        setLoading(true);
        const params = new URLSearchParams(filters);
        axios.get(`http://localhost:8000/connections?${params.toString()}`)
            .then(response => {
                setConnections(response.data || []);
                setLoading(false);
            })
            .catch(error => {
                console.error("There was an error fetching the data!", error);
                setLoading(false);
            });
    };

    useEffect(() => {
        fetchConnections();
    }, []);

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFilters({
            ...filters,
            [name]: value
        });
    };

    const handleFilterSubmit = (e) => {
        e.preventDefault();
        fetchConnections();
    };

    return (
        <Container className="container">
            <Typography variant="h4" component="h1" gutterBottom className="title">
                Load Balancer Connections
            </Typography>
            <form onSubmit={handleFilterSubmit} className="filter-form">
                <TextField
                    label="Client IP"
                    name="client_ip"
                    value={filters.client_ip}
                    onChange={handleInputChange}
                    className="filter-input"
                    variant="outlined"
                    size="small"
                />
                <TextField
                    label="Server URL"
                    name="server_url"
                    value={filters.server_url}
                    onChange={handleInputChange}
                    className="filter-input"
                    variant="outlined"
                    size="small"
                />
                <TextField
                    label="Start Date"
                    name="start_date"
                    type="date"
                    value={filters.start_date}
                    onChange={handleInputChange}
                    className="filter-input"
                    variant="outlined"
                    size="small"
                    InputLabelProps={{
                        shrink: true,
                    }}
                />
                <TextField
                    label="End Date"
                    name="end_date"
                    type="date"
                    value={filters.end_date}
                    onChange={handleInputChange}
                    className="filter-input"
                    variant="outlined"
                    size="small"
                    InputLabelProps={{
                        shrink: true,
                    }}
                />
                <Button type="submit" variant="contained" color="primary" className="filter-button">
                    Filter
                </Button>
            </form>
            {loading ? (
                <div className="loading-spinner">
                    <CircularProgress />
                </div>
            ) : (
                <TableContainer component={Paper} className="table-container">
                    <Table>
                        <TableHead className="table-header">
                            <TableRow>
                                <TableCell className="table-cell">Client IP</TableCell>
                                <TableCell className="table-cell">Server URL</TableCell>
                                <TableCell className="table-cell">Timestamp</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {connections.map((conn, index) => (
                                <TableRow key={index}>
                                    <TableCell>{conn.client_ip}</TableCell>
                                    <TableCell>{conn.server_url}</TableCell>
                                    <TableCell>{conn.timestamp}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            )}
        </Container>
    );
}

export default Metricstcp;
