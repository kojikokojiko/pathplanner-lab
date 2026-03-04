import { useForm } from 'react-hook-form';
import { useNavigate, Link } from 'react-router-dom';
import { useState } from 'react';
import { authApi } from '../api/auth';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import { isCognitoEnabled, startLoginFlow } from '../lib/cognito';

interface FormData {
  email: string;
  password: string;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormData>();

  const onSubmit = async (data: FormData) => {
    setError('');
    try {
      const res = await authApi.login(data.email, data.password);
      localStorage.setItem('token', res.token);
      localStorage.setItem('user', JSON.stringify(res.user));
      navigate('/maps');
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string } } };
      setError(e.response?.data?.detail ?? 'Login failed');
    }
  };

  if (isCognitoEnabled) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <div className="bg-gray-800 rounded-xl border border-gray-700 p-8 w-full max-w-sm text-center">
          <h1 className="text-2xl font-bold text-white mb-2">PathPlanner Lab</h1>
          <p className="text-gray-400 text-sm mb-6">Sign in with your account to continue.</p>
          <Button className="w-full" onClick={() => startLoginFlow()}>
            Sign in with AWS Cognito
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center">
      <div className="bg-gray-800 rounded-xl border border-gray-700 p-8 w-full max-w-sm">
        <h1 className="text-2xl font-bold text-white mb-6">PathPlanner Lab</h1>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            error={errors.email?.message}
            {...register('email', { required: 'Email is required' })}
          />
          <Input
            label="Password"
            type="password"
            placeholder="••••••"
            error={errors.password?.message}
            {...register('password', { required: 'Password is required' })}
          />
          {error && <p className="text-sm text-red-400">{error}</p>}
          <Button type="submit" loading={isSubmitting} className="w-full mt-2">
            Sign In
          </Button>
        </form>
        <p className="text-sm text-gray-400 mt-4 text-center">
          No account?{' '}
          <Link to="/register" className="text-indigo-400 hover:text-indigo-300">
            Register
          </Link>
        </p>
      </div>
    </div>
  );
}
