import { type NextRequest } from 'next/server'

export async function GET(request: NextRequest) {
  // const res = await fetch('https://data.mongodb-api.com/...', {
  //   headers: {
  //     'Content-Type': 'application/json',
  //     'API-Key': process.env.DATA_API_KEY,
  //   },
  // })

  // const data = await res.json()

  console.log(request)

  return Response.json({ data: 'hello airat', request: request.body })
}

const fakeUser = {
  login: 'hui@gmail.com',
  password: 'password',
}

// Login Function
export async function POST(request: NextRequest) {
  const data = await request.json();

  const login = data.login
  const password = data.password

  if (fakeUser.login === login && fakeUser.password === password)
    return Response.json({ GOOD: 'YES', data: 'post hello airat', requestData: data, login, password })

  else 
    return Response.json({ GOOD: 'NO, SOSI DICK', data: 'post hello airat', requestData: data, login, password })

}