const { name } = JSON.parse(process.env.TASK_PAYLOAD)

function sayHello(name) {
  if (name === 'petia') {
    throw new Error('Incorrect name')
  }
  return `hello ${name}`
}

try {
  console.log(JSON.stringify({
    text: sayHello(name)
  }))
} catch (e) {
  console.error(JSON.stringify({
    cause: e.stack,
    error: 'ArgumentException'
  }))
  process.exit(1)
}

