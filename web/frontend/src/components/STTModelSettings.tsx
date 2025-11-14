import { useState, useEffect } from "react";
import { Button } from "./ui/button";
import { Plus, Pencil, Trash2, Check, X } from "lucide-react";
import { STTModelDialog, type STTModel } from "./STTModelDialog";
import { useAuth } from "../contexts/AuthContext";

export function STTModelSettings() {
	const { getAuthHeaders } = useAuth();
	const [models, setModels] = useState<STTModel[]>([]);
	const [loading, setLoading] = useState(true);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [editingModel, setEditingModel] = useState<STTModel | null>(null);

	useEffect(() => {
		fetchModels();
	}, []);

	const fetchModels = async () => {
		try {
			const response = await fetch("/api/v1/admin/stt-models", {
				headers: getAuthHeaders(),
			});

			if (response.ok) {
				const data = await response.json();
				setModels(data);
			} else {
				console.error("Failed to fetch STT models");
			}
		} catch (error) {
			console.error("Error fetching STT models:", error);
		} finally {
			setLoading(false);
		}
	};

	const handleCreate = () => {
		setEditingModel(null);
		setDialogOpen(true);
	};

	const handleEdit = (model: STTModel) => {
		setEditingModel(model);
		setDialogOpen(true);
	};

	const handleDelete = async (id: number) => {
		if (!confirm("Are you sure you want to delete this STT model?")) {
			return;
		}

		try {
			const response = await fetch(`/api/v1/admin/stt-models/${id}`, {
				method: "DELETE",
				headers: getAuthHeaders(),
			});

			if (response.ok) {
				await fetchModels();
			} else {
				const error = await response.json();
				alert(error.error || "Failed to delete STT model");
			}
		} catch (error) {
			console.error("Error deleting STT model:", error);
			alert("Failed to delete STT model");
		}
	};

	const handleSave = async () => {
		await fetchModels();
		setDialogOpen(false);
		setEditingModel(null);
	};

	const getTypeLabel = (type: string) => {
		switch (type) {
			case "whisperx":
				return "WhisperX (Local)";
			case "openai":
				return "OpenAI API";
			case "custom":
				return "Custom API";
			default:
				return type;
		}
	};

	if (loading) {
		return <div className="text-gray-600 dark:text-gray-400">Loading STT models...</div>;
	}

	return (
		<div className="bg-gray-50 dark:bg-gray-700/50 rounded-xl p-4 sm:p-6">
			<div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 sm:gap-0 mb-4">
				<div>
					<h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">
						STT Models
					</h3>
					<p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
						Manage Speech-to-Text models available for transcription
					</p>
				</div>
				<Button
					onClick={handleCreate}
					className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
				>
					<Plus className="h-4 w-4" /> Add Model
				</Button>
			</div>

			{models.length === 0 ? (
				<div className="text-center py-8 text-gray-600 dark:text-gray-400">
					No STT models configured. Add one to get started.
				</div>
			) : (
				<div className="overflow-x-auto">
					<table className="w-full">
						<thead>
							<tr className="border-b border-gray-200 dark:border-gray-700">
								<th className="text-left py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Name
								</th>
								<th className="text-left py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Type
								</th>
								<th className="text-left py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Model
								</th>
								<th className="text-center py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Active
								</th>
								<th className="text-center py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Default
								</th>
								<th className="text-right py-3 px-4 text-sm font-medium text-gray-700 dark:text-gray-300">
									Actions
								</th>
							</tr>
						</thead>
						<tbody>
							{models.map((model) => (
								<tr
									key={model.id}
									className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700/30"
								>
									<td className="py-3 px-4 text-sm text-gray-900 dark:text-gray-100">
										{model.name}
									</td>
									<td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400">
										{getTypeLabel(model.type)}
									</td>
									<td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400">
										{model.model_name || "-"}
									</td>
									<td className="py-3 px-4 text-center">
										{model.is_active ? (
											<Check className="h-4 w-4 text-green-500 mx-auto" />
										) : (
											<X className="h-4 w-4 text-gray-400 mx-auto" />
										)}
									</td>
									<td className="py-3 px-4 text-center">
										{model.is_default ? (
											<Check className="h-4 w-4 text-blue-500 mx-auto" />
										) : (
											<X className="h-4 w-4 text-gray-400 mx-auto" />
										)}
									</td>
									<td className="py-3 px-4 text-right">
										<div className="flex items-center justify-end gap-2">
											<Button
												variant="ghost"
												size="sm"
												onClick={() => handleEdit(model)}
												className="text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-900/20"
											>
												<Pencil className="h-4 w-4" />
											</Button>
											<Button
												variant="ghost"
												size="sm"
												onClick={() => model.id && handleDelete(model.id)}
												disabled={model.is_default}
												className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 disabled:cursor-not-allowed"
											>
												<Trash2 className="h-4 w-4" />
											</Button>
										</div>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}

			<div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
				<p className="flex items-center gap-2">
					<Check className="h-4 w-4 text-blue-500" />
					Default model is automatically selected for new transcriptions
				</p>
				<p className="flex items-center gap-2 mt-1">
					💡 You cannot delete the default model. Set another model as default first.
				</p>
			</div>

			<STTModelDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				initial={editingModel}
				onSave={handleSave}
			/>
		</div>
	);
}
