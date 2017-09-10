import Home from './Home'
import UsersSignIn from './users/SignIn'

export default {
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/users/sign-in',
      name: 'users.sign-in',
      component: UsersSignIn
    }
  ]
}
