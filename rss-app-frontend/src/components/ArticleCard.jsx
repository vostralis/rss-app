import React from 'react';

// Helper function to format the time
const formatTime = (isoString) => {
    if (!isoString) {
        return '';
    }
    try {
        const date = new Date(isoString);
        // Use toLocaleTimeString for locale-aware time formatting
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch (e) {
        return ''; // Return empty if the date is invalid
    }
};

const ArticleCard = ({ article, isFavorite, onToggleFavorite }) => {
    // Safely extract hostname from a URL
    const getHostname = (url) => {
        try {
            return new URL(url).hostname;
        } catch (e) {
            return url;
        }
    };

    return (
        <div className="article-card">
            <div className="article-meta">
                <span className="article-source">{getHostname(article.feedSourceUrl)}</span>
                <span className="article-time">{formatTime(article.publishedAt)}</span>
            </div>
            <h2>{article.title}</h2>
            <p>{article.content}</p>
            
            <div className="card-actions">
                <a href={article.link} target="_blank" rel="noopener noreferrer" className="read-more-link">
                    Read More
                </a>
                {onToggleFavorite && (
                    <button onClick={() => onToggleFavorite(article.id)} className="favorite-btn">
                        {isFavorite ? 'Unfavorite' : 'Favorite'}
                    </button>
                )}
            </div>
        </div>
    );
};

export default ArticleCard;