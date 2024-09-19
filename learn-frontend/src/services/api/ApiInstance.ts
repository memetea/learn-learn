import { Api } from '@/services/api/Api';

const api = new Api({
  baseUrl: 'http://localhost:8081',  // Adjust the base URL if necessary
});

export default api;