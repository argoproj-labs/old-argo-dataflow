module.exports = async function (messageBuf, context) {
  const msg = messageBuf.toString('utf8')
  console.log('Got message', msg)
  return Buffer.from('hi ' + msg)
}