module.exports = async function (messageBuf, context) {
  const msg = messageBuf.toString('utf8')
  return Buffer.from('hi ' + msg)
}