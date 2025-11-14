import { useEffect, useState } from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "./ui/dialog";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "./ui/select";
import { useAuth } from "../contexts/AuthContext";

export type STTModelType = "whisperx" | "openai" | "custom";

export interface STTModel {
	id?: number;
	name: string;
	type: STTModelType;
	base_url?: string;
	model_name?: string;
	is_active: boolean;
	is_default: boolean;
}

interface STTModelDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initial: STTModel | null;
	onSave: (model: STTModel) => Promise<void>;
}

export function STTModelDialog({ open, onOpenChange, initial, onSave }: STTModelDialogProps) {
	const { getAuthHeaders } = useAuth();
	const [formData, setFormData] = useState<STTModel>({
		name: "",
		type: "whisperx",
		is_active: true,
		is_default: false,
	});
	const [apiKey, setApiKey] = useState("");
	const [saving, setSaving] = useState(false);

	useEffect(() => {
		if (initial) {
			setFormData(initial);
			setApiKey(""); // Don't pre-fill API key for security
		} else {
			setFormData({
				name: "",
				type: "whisperx",
				is_active: true,
				is_default: false,
			});
			setApiKey("");
		}
	}, [initial, open]);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setSaving(true);

		try {
			const payload: any = {
				name: formData.name,
				type: formData.type,
				is_active: formData.is_active,
				is_default: formData.is_default,
			};

			// Add type-specific fields
			if (formData.type === "openai" || formData.type === "custom") {
				payload.base_url = formData.base_url;
				payload.model_name = formData.model_name;
				if (apiKey) {
					payload.api_key = apiKey;
				}
			}

			const url = initial?.id
				? `/api/v1/admin/stt-models/${initial.id}`
				: "/api/v1/admin/stt-models";
			const method = initial?.id ? "PUT" : "POST";

			const response = await fetch(url, {
				method,
				headers: {
					...getAuthHeaders(),
					"Content-Type": "application/json",
				},
				body: JSON.stringify(payload),
			});

			if (!response.ok) {
				const error = await response.json();
				alert(error.error || "Failed to save STT model");
				return;
			}

			await onSave(formData);
			onOpenChange(false);
		} catch (error) {
			console.error("Error saving STT model:", error);
			alert("Failed to save STT model");
		} finally {
			setSaving(false);
		}
	};

	const isFormValid = () => {
		if (!formData.name.trim()) return false;
		if (formData.type === "openai" || formData.type === "custom") {
			if (!formData.base_url?.trim()) return false;
			if (!formData.model_name?.trim()) return false;
			// API key is optional for updates
			if (!initial?.id && !apiKey.trim()) return false;
		}
		return true;
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[500px] bg-white dark:bg-gray-800">
				<form onSubmit={handleSubmit}>
					<DialogHeader>
						<DialogTitle className="text-gray-900 dark:text-gray-100">
							{initial ? "Edit STT Model" : "Add STT Model"}
						</DialogTitle>
						<DialogDescription className="text-gray-600 dark:text-gray-400">
							Configure a Speech-to-Text model for transcription.
						</DialogDescription>
					</DialogHeader>

					<div className="grid gap-4 py-4">
						{/* Model Name */}
						<div className="grid gap-2">
							<Label htmlFor="name" className="text-gray-700 dark:text-gray-300">
								Name
							</Label>
							<Input
								id="name"
								value={formData.name}
								onChange={(e) => setFormData({ ...formData, name: e.target.value })}
								placeholder="e.g., OpenAI Whisper API"
								className="bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100"
							/>
						</div>

						{/* Model Type */}
						<div className="grid gap-2">
							<Label htmlFor="type" className="text-gray-700 dark:text-gray-300">
								Type
							</Label>
							<Select
								value={formData.type}
								onValueChange={(value: STTModelType) => setFormData({ ...formData, type: value })}
							>
								<SelectTrigger className="bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100">
									<SelectValue />
								</SelectTrigger>
								<SelectContent className="bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-700">
									<SelectItem value="whisperx">WhisperX (Local)</SelectItem>
									<SelectItem value="openai">OpenAI Whisper API</SelectItem>
									<SelectItem value="custom">Custom API</SelectItem>
								</SelectContent>
							</Select>
						</div>

						{/* API-based model fields */}
						{(formData.type === "openai" || formData.type === "custom") && (
							<>
								<div className="grid gap-2">
									<Label htmlFor="base_url" className="text-gray-700 dark:text-gray-300">
										Base URL
									</Label>
									<Input
										id="base_url"
										value={formData.base_url || ""}
										onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
										placeholder={
											formData.type === "openai"
												? "https://api.openai.com/v1"
												: "https://your-api.com/v1"
										}
										className="bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100"
									/>
								</div>

								<div className="grid gap-2">
									<Label htmlFor="model_name" className="text-gray-700 dark:text-gray-300">
										Model Name
									</Label>
									<Input
										id="model_name"
										value={formData.model_name || ""}
										onChange={(e) => setFormData({ ...formData, model_name: e.target.value })}
										placeholder={formData.type === "openai" ? "whisper-1" : "whisper"}
										className="bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100"
									/>
								</div>

								<div className="grid gap-2">
									<Label htmlFor="api_key" className="text-gray-700 dark:text-gray-300">
										API Key {initial && "(leave empty to keep current)"}
									</Label>
									<Input
										id="api_key"
										type="password"
										value={apiKey}
										onChange={(e) => setApiKey(e.target.value)}
										placeholder="sk-..."
										className="bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-gray-100"
									/>
								</div>
							</>
						)}

						{/* Active Checkbox */}
						<div className="flex items-center gap-2">
							<input
								type="checkbox"
								id="is_active"
								checked={formData.is_active}
								onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
								className="w-4 h-4 rounded border-gray-300 dark:border-gray-600"
							/>
							<Label htmlFor="is_active" className="text-gray-700 dark:text-gray-300">
								Active (available for transcription)
							</Label>
						</div>

						{/* Default Checkbox */}
						<div className="flex items-center gap-2">
							<input
								type="checkbox"
								id="is_default"
								checked={formData.is_default}
								onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
								className="w-4 h-4 rounded border-gray-300 dark:border-gray-600"
							/>
							<Label htmlFor="is_default" className="text-gray-700 dark:text-gray-300">
								Default model
							</Label>
						</div>
					</div>

					<DialogFooter>
						<Button
							type="button"
							variant="outline"
							onClick={() => onOpenChange(false)}
							disabled={saving}
							className="border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300"
						>
							Cancel
						</Button>
						<Button type="submit" disabled={!isFormValid() || saving}>
							{saving ? "Saving..." : initial ? "Update" : "Create"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
