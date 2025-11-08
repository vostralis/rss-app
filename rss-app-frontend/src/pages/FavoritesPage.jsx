import React from 'react';
import { useAppContext } from '../context/AppContext';
import ArticleCard from '../components/ArticleCard';

const FavoritesPage = () => {
    const { articles, favoriteIds } = useAppContext();
    const favoriteArticles = articles.filter(article => favoriteIds.includes(article.id));

    return (
        <div className="page-content">
            <h1>Favorite Articles</h1>
            <div id="favorites-container">
                {favoriteArticles.length === 0 ? (
                    <p>You have no favorite articles yet.</p>
                ) : (
                    favoriteArticles.map(article => (
                        <ArticleCard 
                            key={article.id}
                            article={article}
                        />
                    ))
                )}
            </div>
        </div>
    );
};

export default FavoritesPage;