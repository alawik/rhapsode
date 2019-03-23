'use strict';

const express = require('express');
const bodyParser = require("body-parser");

var func = require('./func.js');

const PORT = 8080;
const HOST = '0.0.0.0';

const app = express();

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());

app.get('/', (req, res) => {
  return res.send('Received a GET HTTP method');
});

app.post('/', (req, res) => {
    //var input = req.body;
    //var math = parseInt(input.a, 10) + parseInt(input.b, 10);
    //var math = func(input);
    //return res.send(func(req.body).toString());
    //return res.end(func(req.body).toString());
    res.end(func(req.body).toString());
});

app.listen(PORT, () =>
  //console.log(`Example app listening on port ${process.env.PORT}!`),
  console.log(`Running on http://${HOST}:${PORT}`),
);
