import React from 'react';
import { useAppContext } from '../context/AppContext';
import ArticleCard from '../components/ArticleCard';

const NewsPage = () => {
    const { articles, favoriteIds, toggleFavorite, loading } = useAppContext();

    if (loading) {
        return <div className="page-content"><h1>Loading articles...</h1></div>;
    }

    return (
        <div className="page-content">
            <h1>All News</h1>
            <div id="news-list">
                {articles.map(article => (
                    <ArticleCard 
                        key={article.id}
                        article={article}
                        isFavorite={favoriteIds.includes(article.id)}
                        onToggleFavorite={toggleFavorite}
                    />
                ))}
            </div>
        </div>
    );
};

export default NewsPage;