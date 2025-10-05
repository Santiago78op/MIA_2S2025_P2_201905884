export const formatDate = (date: Date): string => {
    return date.toLocaleDateString();
};

export const fetchData = async (url: string): Promise<any> => {
    const response = await fetch(url);
    if (!response.ok) {
        throw new Error('Network response was not ok');
    }
    return response.json();
};