import { useEffect, useState } from 'react';
import {   ModelsQuestionBank } from '@/services/api/Api';
import api from '@/services/api/ApiInstance';
import { Button } from "@/components/ui/button";


function QuestionBankList() {
  const [questionBanks, setQuestionBanks] = useState<ModelsQuestionBank[]>([]);

  useEffect(() => {
    api.questionBanks.questionBanksList()
      .then(response => {
        setQuestionBanks(response.data);
      })
      .catch(error => {
        console.error("Error fetching question banks:", error);
      });
  }, []);

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Question Banks</h1>
      <ul>
        {questionBanks.map(bank => (
          <li key={bank.id} className="mb-2">
            <Button variant="default">{bank.name}</Button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default QuestionBankList;
