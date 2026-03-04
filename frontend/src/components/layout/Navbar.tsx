import { Link, useNavigate } from 'react-router-dom';
import { Map, FlaskConical, GitCompare, LogOut } from 'lucide-react';

export default function Navbar() {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  return (
    <nav className="bg-gray-800 border-b border-gray-700">
      <div className="max-w-7xl mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center gap-8">
            <Link to="/" className="text-indigo-400 font-bold text-lg tracking-tight">
              PathPlanner Lab
            </Link>
            <div className="flex items-center gap-4">
              <Link
                to="/maps"
                className="flex items-center gap-2 text-gray-300 hover:text-white transition-colors text-sm"
              >
                <Map size={16} />
                Maps
              </Link>
              <Link
                to="/experiments"
                className="flex items-center gap-2 text-gray-300 hover:text-white transition-colors text-sm"
              >
                <FlaskConical size={16} />
                Experiments
              </Link>
              <Link
                to="/compare"
                className="flex items-center gap-2 text-gray-300 hover:text-white transition-colors text-sm"
              >
                <GitCompare size={16} />
                Compare
              </Link>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors text-sm"
          >
            <LogOut size={16} />
            Logout
          </button>
        </div>
      </div>
    </nav>
  );
}
