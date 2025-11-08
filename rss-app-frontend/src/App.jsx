import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import FeedsPage from './pages/FeedsPage';
import NewsPage from './pages/NewsPage';
import FavoritesPage from './pages/FavoritesPage';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<FeedsPage />} />
        <Route path="news" element={<NewsPage />} />
        <Route path="favorites" element={<FavoritesPage />} />
      </Route>
    </Routes>
  );
}

export default App;