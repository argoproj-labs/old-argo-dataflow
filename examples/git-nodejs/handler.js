module.exports = async function (messageBuf, context) {
  const msg = messageBuf.toString('utf8')
  console.log('Got message', msg)
  await new Promise(r => setTimeout(r, 10000))
  return Buffer.from('hi ' + msg)
}