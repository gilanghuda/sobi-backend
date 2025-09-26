CREATE TABLE history_education (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),            
    user_id UUID NOT NULL,                                       
    education_id UUID NOT NULL,                                                       
    CONSTRAINT fk_history_user FOREIGN KEY (user_id) REFERENCES users(uid) ON DELETE CASCADE,
    CONSTRAINT fk_history_education FOREIGN KEY (education_id) REFERENCES educations(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_education UNIQUE (user_id, education_id) 
);
