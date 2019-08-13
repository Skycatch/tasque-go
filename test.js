const values = JSON.parse(process.env.TASK_PAYLOAD)
console.log('log message');
console.log('log message');
console.log('log message');
console.log('log message');
if (values.name === 'table') {
  console.error('incorrect name')
  console.error('-=error=-')
  console.error(JSON.stringify({ error: 'ArgumentException' }));
  process.exit(1)
} else {
  console.log('-=result=-');
  console.log(JSON.stringify({ response: `Hello ${values.name}` }));
}
console.log('log message');