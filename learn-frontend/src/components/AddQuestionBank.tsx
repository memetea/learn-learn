import React, { useState } from 'react';
import {  ModelsQuestionBank } from '@/services/api/Api';
import api from '@/services/api/ApiInstance';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';


function AddQuestionBank({ onAdd }: { onAdd: () => void }) {
  const [name, setName] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const newQuestionBank: ModelsQuestionBank = { name };
      await api.questionBanks.questionBanksCreate(newQuestionBank);
      setName(''); // Clear the input field
      onAdd(); // Notify parent component to update the list
    } catch (error) {
      console.error("Failed to add question bank:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center space-x-2 mb-4">
      <Input
        type="text"
        placeholder="Enter question bank name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        required
      />
      <Button type="submit">Add Question Bank</Button>
    </form>
  );
}

export default AddQuestionBank;
