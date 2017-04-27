function main(req, res) {
  const say = (req.body && req.body.say)
  res.send({ say: `${say} I hope you enjoyed this talk at QconSP! Thank you ;)` })
}
