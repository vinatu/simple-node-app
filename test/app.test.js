const request = require('supertest');
const express = require('express');

const app = express();
app.get('/', (req, res) => {
  res.status(200).send('Hello, World!');
});

describe('GET /', () => {
  it('should return Hello, World!', async () => {
    const res = await request(app).get('/');
    expect(res.statusCode).toEqual(200);
    expect(res.text).toBe('Hello, World!');
  });
});
