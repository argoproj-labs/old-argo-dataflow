const http = require('http')

const host = 'localhost'
const port = 8080

let defaultHandler = null

async function getHandler (req, res) {
  res.writeHead(204)
  res.end()
}

async function postHandler (req, res) {
  try {
    let contentBuffer = []
    let totalBytesInBuffer = 0

    req.on('data', chunk => {
      contentBuffer.push(chunk)
      totalBytesInBuffer += chunk.length
    })

    req.on('end', async function () {
      contentBuffer = Buffer.concat(contentBuffer, totalBytesInBuffer)

      try {
        const out = await defaultHandler(contentBuffer, {})

        if (out) {
          res.writeHead(201)
          res.end(out)
        } else {
          res.writeHead(204)
          res.end()
        }
      } catch (err) {
        console.error('Handler failed to process the message', err)
        res.writeHead(500)
        res.end(err.message)
      }
    })
  } catch (err) {
    console.log('Got an error proceesing a request!', err)
    res.writeHead(500)
    res.end()
  }
}

async function requestListener (req, res) {
  if (req.method === 'GET') {
    return getHandler(req, res)
  }

  if (req.method === 'POST') {
    return postHandler(req, res)
  }

  res.writeHead(200)
  res.end('OK')
};

async function start (handler) {
  defaultHandler = handler
  const server = http.createServer(requestListener)
  server.listen(port, host, () => {
    console.log(`Server is running on http://${host}:${port} and pid ${process.pid}`)
  })

  process.on('SIGTERM', function () {
    console.log('Got SIGTERM, shutting down server.')
    server.close(function () {
      console.log('Server shut down, terminating.')
      process.exit(0)
    })
  })
}

module.exports = { start }
