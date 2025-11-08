import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';

const Layout = () => {
  return (
    <>
      <nav>
        <NavLink to="/">Feeds</NavLink>
        <NavLink to="/news">All News</NavLink>
        <NavLink to="/favorites">Favorites</NavLink>
      </nav>
      <main>
        <Outlet />
      </main>
    </>
  );
};

export default Layout;