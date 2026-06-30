import axios from 'axios'

export default {
  getTopAndFeaturedArticles: () => {
    return axios.get('/api/articles/topAndFeatured')
  },
  getArticles: (params: any) => {
    return axios.get('/api/articles/all', { params: params })
  },
  getArticlesByCategoryId: (params: any) => {
    return axios.get('/api/articles/categoryId', { params: params })
  },
  getArticeById: (articleId: any) => {
    return axios.get('/api/articles/' + articleId)
  },
  getAllCategories: () => {
    return axios.get('/api/categories/all')
  },
  getAllTags: () => {
    return axios.get('/api/tags/all')
  },
  getTopTenTags: () => {
    return axios.get('/api/tags/topTen')
  },
  getArticlesByTagId: (params: any) => {
    return axios.get('/api/articles/tagId', { params: params })
  },
  getAllArchives: (params: any) => {
    return axios.get('/api/archives/all', { params: params })
  },
  login: (params: any) => {
    return axios.post('/api/users/login', params)
  },
  saveComment: (params: any) => {
    return axios.post('/api/comments/save', params)
  },
  getComments: (params: any) => {
    return axios.get('/api/comments', { params: params })
  },
  getTopSixComments: () => {
    return axios.get('/api/comments/topSix')
  },
  getAbout: () => {
    return axios.get('/api/about')
  },
  getFriendLink: () => {
    return axios.get('/api/links')
  },
  submitUserInfo: (params: any) => {
    return axios.put('/api/users/info', params)
  },
  getUserInfoById: (id: any) => {
    return axios.get('/api/users/info/' + id)
  },
  updateUserSubscribe: (params: any) => {
    return axios.put('/api/users/subscribe', params)
  },
  sendValidationCode: (username: any) => {
    return axios.get('/api/users/code', {
      params: {
        username: username
      }
    })
  },
  bindingEmail: (params: any) => {
    return axios.put('/api/users/email', params)
  },
  register: (params: any) => {
    return axios.post('/api/users/register', params)
  },
  searchArticles: (params: any) => {
    return axios.get('/api/articles/search', {
      params: params
    })
  },
  getAlbums: () => {
    return axios.get('/api/photos/albums')
  },
  getPhotosByAlbumId: (albumId: any, params: any) => {
    return axios.get('/api/albums/' + albumId + '/photos', {
      params: params
    })
  },
  getWebsiteConfig: () => {
    return axios.get('/api/website/config')
  },
  qqLogin: (params: any) => {
    return axios.post('/api/users/oauth/qq', params)
  },
  report: () => {
    return axios.post('/api/report').catch((e) => {
      console.warn('[Report] 访问上报失败:', e.message)
      return Promise.resolve(null)
    })
  },
  getTalks: (params: any) => {
    return axios.get('/api/talks', {
      params: params
    })
  },
  getTalkById: (id: any) => {
    return axios.get('/api/talks/' + id)
  },
  logout: () => {
    return axios.post('/api/users/logout')
  },
  getRepliesByCommentId: (commentId: any) => {
    return axios.get(`/api/comments/${commentId}/replies`)
  },
  updatePassword: (params: any) => {
    return axios.put('/api/users/password', params)
  },
  accessArticle: (params: any) => {
    return axios.post('/api/articles/access', params)
  },
  getPortfoliosFeatured: () => {
    return axios.get('/api/portfolios/featured')
  },
  getPortfolios: (params: any) => {
    return axios.get('/api/portfolios', { params: params })
  }
}
