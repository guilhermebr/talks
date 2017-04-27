function main(req, res) {
  const name = (req.body && req.body.name) || "World"
  res.send({ say: `Hello ${name}!` })
}
