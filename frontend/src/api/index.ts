import axios from 'axios'
const baseUrl = '/api/v1/'
// const baseUrl = 'http://localhost:3000/api/'

const api = axios.create({
  baseURL: baseUrl,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded'
  },
  transformRequest: [
    (data) => {
      if (data && typeof data === 'object') {
        const params = new URLSearchParams()
        for (const key in data) {
          if (data[key] != null) {
            params.append(key, data[key])
          }
        }
        return params.toString()
      }
      return data
    }
  ]
})
export const _api = axios.create({
  baseURL: baseUrl,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
})

export default api
