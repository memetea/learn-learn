import React, { useState } from 'react';
import QuestionBankList from './components/QuestionBankList';
import AddQuestionBank from './components/AddQuestionBank';

function App() {
  const [refresh, setRefresh] = useState(false);

  // Function to trigger a refresh of the QuestionBank list
  const handleRefresh = () => {
    setRefresh(!refresh);
  };

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <h1 className="text-3xl font-bold text-center mb-6">Question Bank Manager</h1>
      <AddQuestionBank onAdd={handleRefresh} />
      <QuestionBankList key={`refresh-${refresh}`} />
    </div>
  );
}

export default App;
