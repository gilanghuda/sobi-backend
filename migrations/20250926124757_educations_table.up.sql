CREATE TABLE educations (
    id UUID PRIMARY KEY,         
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255),   
    video_url TEXT,                
    duration VARCHAR(10),        
    author VARCHAR(255),                     
    description TEXT,        
    created_at TIMESTAMP DEFAULT now()
);
